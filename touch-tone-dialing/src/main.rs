use std::collections::HashMap;

use env_logger::Env;
use hound::WavSpec;
use serde_json::json;

const LOW_FREQS: [f64; 4] = [697_f64, 770_f64, 852_f64, 941_f64];
const HIGH_FREQS: [f64; 4] = [1209_f64, 1336_f64, 1477_f64, 1633_f64];
const PI: f64 = std::f64::consts::PI;
const DTMF_CHARS: [[char; 4]; 4] = [
    ['1', '2', '3', 'A'],
    ['4', '5', '6', 'B'],
    ['7', '8', '9', 'C'],
    ['*', '0', '#', 'D'],
];

fn map_freqs_to_char(low_freq_idx: usize, high_freq_idx: usize) -> char {
    DTMF_CHARS[low_freq_idx][high_freq_idx]
}

// 戈泽尔算法实现
// 计算单个频率在样本块中的能量（幅度的平方）
fn goertzel(samples: &[f64], target_freq: f64, sample_rate: u32) -> f64 {
    let n = samples.len() as f64;
    let k = (0.5 + n * target_freq / sample_rate as f64) as i32;
    let omega = (2.0 * PI * k as f64) / n;
    let cosine = omega.cos();
    let coeff = 2.0 * cosine;

    let mut s1 = 0.0;
    let mut s2 = 0.0;

    for &sample in samples {
        let s0 = sample + coeff * s1 - s2;
        s2 = s1;
        s1 = s0;
    }

    s1 * s1 + s2 * s2 - coeff * s1 * s2
}

fn decode_dtmf(samples: Vec<i16>, spec: WavSpec) -> String {
    let samples_f64: Vec<f64> = samples.iter().map(|&s| s as f64).collect();
    // DTMF 音调通常持续至少 40ms，这里我们选择一个合适的块大小
    // 对于 4000Hz 采样率，103 个样本约等于 25ms，这是一个常见的大小
    let block_size = 103;
    let sample_rate = spec.sample_rate;

    // 这是一个经验值，用于过滤噪音。
    let threshold = 1.0e9;
    let mut decoded_sequence = String::new();
    let mut last_char = '\0'; // 用于防止长音被重复解码
    let mut in_tone = false; // 标记当前是否在音调中
    for block in samples_f64.chunks(block_size) {
        if block.len() < block_size {
            continue;
        }
        let low_magnitudes: Vec<f64> = LOW_FREQS
            .iter()
            .map(|&f| goertzel(block, f, sample_rate))
            .collect();

        let high_magnitudes: Vec<f64> = HIGH_FREQS
            .iter()
            .map(|&f| goertzel(block, f, sample_rate))
            .collect();

        // 找到低频和高频组中能量最强的频率
        let (low_idx, &low_max) = low_magnitudes
            .iter()
            .enumerate()
            .max_by(|a, b| a.1.partial_cmp(b.1).unwrap())
            .unwrap();

        let (high_idx, &high_max) = high_magnitudes
            .iter()
            .enumerate()
            .max_by(|a, b| a.1.partial_cmp(b.1).unwrap())
            .unwrap();

        // --- 解码和状态管理 ---
        if low_max > threshold && high_max > threshold {
            // 检测到有效音调
            let current_char = map_freqs_to_char(low_idx, high_idx);
            if !in_tone || current_char != last_char {
                decoded_sequence.push(current_char);
                last_char = current_char;
            }
            in_tone = true;
        } else {
            // 检测到静音
            if in_tone {
                // 从音调变为静音，重置 last_char
                // 这使得 "11" (音-静-音) 可以被正确解码
                last_char = '\0';
            }
            in_tone = false;
        }
    }
    return decoded_sequence;
}

fn main() -> anyhow::Result<()> {
    // Logger Config
    env_logger::Builder::from_env(Env::new().default_filter_or("info")).init();
    let response_json = reqwest::blocking::get(
        "https://hackattic.com/challenges/touch_tone_dialing/problem?access_token=aa8e14a982d0994e",
    )?
    .text()?;
    let binding = serde_json::from_str::<HashMap<String, String>>(&response_json)?;
    let wav_url = binding.get("wav_url").unwrap();
    let res = reqwest::blocking::get(wav_url)?;

    // std::fs::write("./wavfile.wav", res.bytes()?)?;
    let binding = res.bytes()?;
    let mut reader = hound::WavReader::new(binding.iter().as_slice())?;
    let spec = reader.spec();
    log::info!(
        "Wav channel count {}, sample rate {}, bits per sample {}",
        spec.channels,
        spec.sample_rate,
        spec.bits_per_sample
    );
    let samples: Vec<i16> = reader.samples::<i16>().collect::<Result<Vec<_>, _>>()?;
    let str = decode_dtmf(samples, spec);
    log::info!("Decoded str {}", str);
    let response = reqwest::blocking::Client::new().post(
        "https://hackattic.com/challenges/touch_tone_dialing/solve?access_token=aa8e14a982d0994e",
    ).body(
        json!({
            "sequence": str
        }).to_string()
    ).send()?;
    log::info!("res text {}", response.text().unwrap());
    Ok(())
}
