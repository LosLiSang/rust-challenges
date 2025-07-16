use std::time::{self, Duration};

use env_logger::Env;
use openssl::{
    asn1::{Asn1Integer, Asn1Time},
    base64,
    bn::BigNum,
    nid::Nid,
    pkey::PKey,
    x509::X509Name,
};
use serde_json::Value;

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();

    let base_url = "https://hackattic.com".to_string();
    let json_str = reqwest::get(
        base_url.clone() + "/challenges/tales_of_ssl/problem?access_token=aa8e14a982d0994e",
    )
    .await?
    .text()
    .await?;
    log::info!("receive json {}", json_str);
    let map = serde_json::from_str::<Value>(&json_str)?;
    let private_key = map.get("private_key").unwrap();
    let required_data = map.get("required_data").unwrap();
    let domain = required_data.get("domain").unwrap();
    let serial_number = required_data.get("serial_number").unwrap();
    let country = required_data.get("country").unwrap();
    // log::info!("private key {}", private_key.as_str().unwrap());
    let der_private_key = base64::decode_block(private_key.as_str().unwrap()).unwrap();
    let pkey = PKey::private_key_from_der(der_private_key.as_slice()).unwrap();

    // ...existing code...
    let country_code = match country.as_str().unwrap() {
        "Keeling Islands" => "CC",
        "Cocos Islands" => "CC",
        "Cocos (Keeling) Islands" => "CC",
        "Christmas Island" => "CX",
        "Sint Maarten" => "SX",
        "Saint Martin" => "SX", // 可能的别名
        "St. Maarten" => "SX",  // 可能的别名
        // 其他国家...
        "United States" => "US",
        "United Kingdom" => "GB",
        // 如果已经是2位代码，直接使用
        code if code.len() == 2 => code,
        // 默认情况
        _ => {
            log::warn!("Unknown country: {}, using 'XX'", country.as_str().unwrap());
            "XX"
        }
    };
    // ...existing code...
    let mut name = X509Name::builder().unwrap();
    name.append_entry_by_nid(Nid::COMMONNAME, domain.as_str().unwrap())
        .unwrap();
    name.append_entry_by_nid(Nid::ORGANIZATIONNAME, country.as_str().unwrap())
        .unwrap();
    name.append_entry_by_nid(Nid::COUNTRYNAME, country_code)
        .unwrap();
    let name = name.build();

    let start_time = Asn1Time::from_unix(
        time::SystemTime::now()
            .duration_since(time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64,
    )
    .unwrap();
    let end_time = Asn1Time::from_unix(
        time::SystemTime::now()
            .checked_add(Duration::from_secs(3600 * 24 * 360))
            .unwrap()
            .duration_since(time::UNIX_EPOCH)
            .unwrap()
            .as_secs() as i64,
    )
    .unwrap();
    // 去掉 "0x" 前缀
    let hex_str = if serial_number.as_str().unwrap().starts_with("0x") {
        &serial_number.as_str().unwrap()[2..]
    } else {
        serial_number.as_str().unwrap()
    };
    let serial_number =
        Asn1Integer::from_bn(BigNum::from_hex_str(hex_str).unwrap().as_ref()).unwrap();
    let mut builder = openssl::x509::X509Builder::new().unwrap();
    builder.set_pubkey(&pkey).unwrap();
    builder.set_serial_number(serial_number.as_ref()).unwrap();
    builder.set_issuer_name(&name).unwrap();
    builder.set_subject_name(&name).unwrap();
    builder.set_not_before(&start_time).unwrap();
    builder.set_not_after(&end_time).unwrap();
    builder
        .sign(&pkey, openssl::hash::MessageDigest::sha256())
        .unwrap();
    let cert = builder.build();
    let res = cert.to_der().unwrap();
    let base64_cert = base64::encode_block(res.as_slice());
    // log::info!("base64 cert: {}", base64_cert);
    let res = reqwest::Client::new()
        .post(base_url + "/challenges/tales_of_ssl/solve?access_token=aa8e14a982d0994e")
        .json(&serde_json::json!({
            "certificate": base64_cert
        }))
        .send()
        .await
        .unwrap();

    log::info!("{}", res.text().await.unwrap());
    Ok(())
}
