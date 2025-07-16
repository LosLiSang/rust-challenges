use std::u128;

use env_logger::Env;
use futures_util::{SinkExt, StreamExt};
use tokio::time::Instant;
use tokio_tungstenite::tungstenite::{Message, Utf8Bytes, client::IntoClientRequest};

fn check(duration: u128) -> &'static str {
    match duration {
        600..1400 => return "700",
        1400..1900 => return "1500",
        1900..2400 => return "2000",
        2400..2900 => return "2500",
        2900..3500 => return "3000",
        _ => return "700",
    }
}

async fn send_solition(secret: String) -> anyhow::Result<()> {
    let text = reqwest::Client::new()
        .post("https://hackattic.com/challenges/websocket_chit_chat/solve?access_token=aa8e14a982d0994e")
        .body( format!("{{\"secret\": \"{}\"}}", secret))
        .send().await?.text().await?;
    log::info!("{}", text);
    return Ok(());
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // // Logger Config
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();

    let token_url = "https://hackattic.com/challenges/websocket_chit_chat/problem?access_token=aa8e14a982d0994e";
    let binding = reqwest::get(token_url).await.unwrap().text().await.unwrap();
    let text = binding.as_bytes();
    let token = str::from_utf8(&text[11..text.len() - 2]).unwrap();

    let base_url = String::from("wss://hackattic.com/_/ws/");
    let wss_url = base_url + token;

    let (ws_stream, response) =
        tokio_tungstenite::connect_async(wss_url.into_client_request().unwrap()).await?;
    let mut start = Instant::now();
    log::info!("Connect Success");
    log::info!("Server Response: {}", response.status());
    let (mut writer, mut reader) = ws_stream.split();
    let r_task = tokio::spawn(async move {
        while let Some(msg) = reader.next().await {
            match msg {
                Ok(message) => match message {
                    tokio_tungstenite::tungstenite::Message::Text(utf8_bytes) => {
                        if utf8_bytes.to_string() == "ping!" {
                            let duration = start.elapsed();
                            start = Instant::now();
                            log::info!("Recv Text Msg: {}", utf8_bytes);
                            log::info!("It takes {} mill seconds", duration.as_millis());
                            writer
                                .send(Message::Text(Utf8Bytes::from_static(check(
                                    duration.as_millis(),
                                ))))
                                .await
                                .unwrap();
                        } else if utf8_bytes.to_string() == "good!" {
                        } else if utf8_bytes
                            .to_string()
                            .starts_with("congratulations! the solution to this challenge is \"")
                        {
                            let smsg = utf8_bytes.to_string();
                            let secret = str::from_utf8(
                                &smsg.as_bytes()
                                    ["congratulations! the solution to this challenge is \"".len()
                                        ..smsg.len() - 1],
                            ).unwrap();
                            log::info!("Recv Secret: {}", secret);
                            tokio::spawn(send_solition(String::from(secret))).await.unwrap().unwrap();
                        } else {
                            log::info!("Recv Text Msg: {}", utf8_bytes);
                        }
                    }
                    tokio_tungstenite::tungstenite::Message::Binary(_) => {}
                    tokio_tungstenite::tungstenite::Message::Ping(_) => {}
                    tokio_tungstenite::tungstenite::Message::Pong(_) => {}
                    tokio_tungstenite::tungstenite::Message::Close(_) => {}
                    tokio_tungstenite::tungstenite::Message::Frame(_) => {}
                },
                Err(_) => {}
            }
        }
    });

    r_task.await?;
    Ok(())
}
