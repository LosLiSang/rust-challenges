use std::{
    collections::HashMap,
    io::Read,
    sync::{Arc, Mutex}, time::{self, SystemTime, UNIX_EPOCH},
};

use axum::{
    Json, Router,
    body::Body,
    extract::{Request, State},
    http::HeaderMap,
    middleware::{self, Next},
    response::IntoResponse,
    routing::post,
};
use base64::Engine;
use env_logger::Env;
use hmac::{Hmac, Mac};
use serde::Deserialize;
use serde_json::{Value, json};
use sha2::Sha256;

struct AppState {
    solution: Arc<Mutex<String>>,
    jwt_secret: String,
}

#[axum::debug_handler]
async fn handler(State(map): State<Arc<AppState>>, token: String) -> impl IntoResponse {
    log::info!("token {}", token);
    let mut items: Vec<&[u8]> = token.as_bytes().split(|&x| x == b'.').collect();

    let mut header = String::new();
    let mut payload = String::new();
    let mut origin_sign = String::new();
    base64::prelude::BASE64_URL_SAFE_NO_PAD
        .decode(items[0])
        .unwrap()
        .as_slice()
        .read_to_string(&mut header)
        .unwrap();
    base64::prelude::BASE64_URL_SAFE_NO_PAD
        .decode(items[1])
        .unwrap()
        .as_slice()
        .read_to_string(&mut payload)
        .unwrap();
    items[2].read_to_string(&mut origin_sign).unwrap();
    log::info!("head json: {}", header);
    log::info!("payload json: {}", payload);
    let header_map = serde_json::from_str::<HashMap<String, String>>(&header).unwrap();
    let payload_map = serde_json::from_str::<HashMap<String, Value>>(&payload).unwrap();
    log::info!("header map: {:?}", header_map);
    
    if header_map.get("alg").unwrap() != "HS256" || header_map.get("typ").unwrap() != "JWT" {
        //
    }

    type HmacSha256 = Hmac<Sha256>;
    let mut mac = HmacSha256::new_from_slice(map.jwt_secret.clone().as_bytes()).unwrap();
    mac.update(items[0..2].join(&b'.').as_slice());
    let result = mac.finalize().into_bytes();
    let signature = base64::prelude::BASE64_URL_SAFE_NO_PAD.encode(result);
    log::info!("origin signature {}", origin_sign);
    log::info!("verify signature {}", signature);

    let exp = payload_map.get("exp").unwrap_or(&json!(17521374670u64)).as_u64().unwrap();    
    let nbf = payload_map.get("nbf").unwrap_or(&json!(0)).as_u64().unwrap();    
    let now = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_secs();
    let is_valid_token = now >= nbf && now <= exp;

    if !payload_map.contains_key("append") {
        let solution = { map.solution.clone().lock().unwrap().clone() };
        Json(json!({
            "solution": solution
        }))
    } else {
        let str = payload_map.get("append").unwrap();
        let mut solution = { map.solution.lock().unwrap() };

        if signature == origin_sign && is_valid_token {
            *solution = solution.clone() + &str.as_str().unwrap();
        }
        Json(serde_json::Value::Null)
    }
}


#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();
    let res = reqwest::get(
        "https://hackattic.com/challenges/jotting_jwts/problem?access_token=aa8e14a982d0994e",
    )
    .await?;
    let json = res.text().await?;
    #[derive(Deserialize)]
    struct Response {
        jwt_secret: String,
    }
    let jwt_secret = serde_json::from_str::<Response>(&json)?.jwt_secret;
    log::info!("fetch jwt_secret: {}", jwt_secret);
    // eD&l945vvm6RMNya
    let app_state = Arc::new(AppState {
        solution: Arc::new(Mutex::new("".to_string())),
        jwt_secret,
    });

    let router = Router::default()
        .route("/", post(handler))
        .with_state(app_state.clone());

    tokio::spawn(async {
        let res = reqwest::Client::new()
            .post(
                "https://hackattic.com/challenges/jotting_jwts/solve?access_token=aa8e14a982d0994e",
            )
            .body(
                json!(
                    {"app_url": "http://47.97.56.41:3000/"}
                )
                .to_string(),
            )
            .send()
            .await
            .unwrap()
            .text()
            .await
            .unwrap();

        log::info!("send post success {}", res);
    });
    let url = "0.0.0.0:3000";
    let listener = tokio::net::TcpListener::bind(url).await?;
    log::info!("server start at {}", url);
    axum::serve(listener, router).await?;

    Ok(())
}
