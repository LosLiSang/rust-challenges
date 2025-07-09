use std::collections::HashMap;

use anyhow::anyhow;
use env_logger::Env;
use serde::Deserialize;
use serde_json::Value;

#[derive(Deserialize)]
#[serde(rename_all = "camelCase")]
struct ResponseBody {
    bytes: String,
}



async fn fetch_data(url: &str) -> anyhow::Result<String> {
    let rep = reqwest::get(url)
        .await
        .map_err(|e| anyhow!("error{:}", e))?;
    
    let test = rep
        .text()
        .await
        .map_err(|e| anyhow!("unwarp response bytes error{:}", e))?;

    Ok(test)
}

// 编码后的每4个字节对应编码前的每3个字节
fn base64_decode(base64_str: String) -> anyhow::Result<Vec<u8>> {
    let base64_chars =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/".as_bytes();
    let mut decode_map = HashMap::new();
    for (i, u8) in base64_chars.iter().enumerate() {
        decode_map.entry(u8).or_insert(i as u8);
    }
    let encoded_bytes = base64_str.as_bytes();
    let mut index = 0;
    let mut decode_bytes = vec![];
    while index < encoded_bytes.len() {
        let b0 = decode_map.get(&encoded_bytes[index]).unwrap_or(&0);
        let b1 = decode_map.get(&encoded_bytes[index + 1]).unwrap_or(&0);
        let b2 = decode_map.get(&encoded_bytes[index + 2]).unwrap_or(&0);
        let b3 = decode_map.get(&encoded_bytes[index + 3]).unwrap_or(&0);
        let db1 = (b0 % 64 << 2) + b1 / 16;
        let db2 = (b1 % 16 << 4) + b2 / 4;
        let db3 = (b2 % 4 << 6) + b3;
        decode_bytes.push(db1);
        decode_bytes.push(db2);
        decode_bytes.push(db3);
        index += 4;
    }
    log::info!("{:?}", decode_bytes);

    Ok(decode_bytes)
}

async fn submit(map: &HashMap<&'static str, Value>) -> anyhow::Result<()> {
    let uri = "https://hackattic.com/challenges/help_me_unpack/solve?access_token=aa8e14a982d0994e";
    log::info!("Submitting data: {:?}", map);
    let client = reqwest::Client::new();
    let res = client
        .post(uri)
        .json(map)
        .send()
        .await
        .map_err(|e| anyhow!("err: {e}"))?;
    log::info!("{:?}", res.text().await?);
    Ok(())
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();

    let uri =
        "https://hackattic.com/challenges/help_me_unpack/problem?access_token=aa8e14a982d0994e";
    let body = &fetch_data(uri).await?;
    // let body = "{\"bytes\": \"eR5xgWzB+cVVhwAA5EdhQzwhdIB0bYBAQIBtdIB0ITw=\"}";

    let base64_str = serde_json::from_str::<ResponseBody>(body)
        .map_err(|e| anyhow!("serde entity error {}", e))?
        .bytes;

    log::info!("{:?}", base64_str);
    let decode_bytes = base64_decode(base64_str)?;
    let int = i32::from_le_bytes(decode_bytes[0..4].try_into().unwrap());
    let unsigned_int = u32::from_le_bytes(decode_bytes[4..8].try_into().unwrap());
    let short = i16::from_le_bytes(decode_bytes[8..10].try_into().unwrap());
    let float = f32::from_le_bytes(decode_bytes[12..16].try_into().unwrap());
    let double = f64::from_le_bytes(decode_bytes[16..24].try_into().unwrap());
    let be_double = f64::from_be_bytes(decode_bytes[24..32].try_into().unwrap());

    // [121, 30, 113, 129, 108, 193, 249, 197, 85, 135, 0, 0, 228, 71, 97, 67, 60, 33, 116, 128, 116, 109, 128, 64, 64, 128, 109, 116, 128, 116, 33, 60, 0]
    // 4B int           121, 30, 113, 129
    // 4B unsigned int  108, 193, 249, 197,
    // 2B short         85, 135, 0, 0,
    // 4B float         228, 71, 97, 67,
    // 8B double         60, 33, 116, 128, 116, 109, 128, 64,
    // 8B double big endian     64, 128, 109, 116, 128, 116, 33, 60
    log::info!("{int} {unsigned_int} {short} {float} {double} {be_double}");
    let mut map: HashMap<&'static str, Value> = HashMap::new();
    map.entry("int").or_insert(int.into());
    map.entry("uint").or_insert(unsigned_int.into());
    map.entry("short").or_insert(short.into());
    map.entry("float").or_insert(float.into());
    map.entry("double").or_insert(double.into());
    map.entry("big_endian_double")
        .or_insert(be_double.into());
    submit(&map).await?;
    Ok(())
}
