use serde::{Deserialize, Serialize};
use serde_json::Value;



#[derive(Deserialize, Serialize, Debug)]
struct Credentials {
    user: String,
    name: String,
}

#[derive(Deserialize, Serialize, Debug)]
struct FetchData {
    ignition_key: String,
    trigger_token: String,
    credentials: Credentials
}

const FETCH_URL: &str =
    "https://hackattic.com/challenges/dockerized_solutions/problem?access_token=aa8e14a982d0994e";

async fn fetch_data() -> anyhow::Result<FetchData> {
    let text = reqwest::get(FETCH_URL).await?.text().await?;
    let res = serde_json::from_str::<FetchData>(&text)?;
    Ok(res)
}

async fn push_base() -> anyhow::Result<()> {

    
    Ok(())
}

fn main() {
    let runtime = tokio::runtime::Builder::new_multi_thread()
        .worker_threads(8)
        .enable_io() // 可在runtime中使用异步IO
        .enable_time() // 可在runtime中使用异步计时器(timer)
        .build()
        .unwrap();
    runtime.block_on(async {

        let data = fetch_data().await.unwrap();
        log::info!("Fetch data: {:?}", data);
        


    });
}
