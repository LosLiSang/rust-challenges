# Help me unpack

[https://hackattic.com/challenges/help_me_unpack](网站)

## 要求

连接到问题端点，抓取并解压base64编码的字节包

> The pack contains, always in the following order:
> a regular int (signed), to start off
> an unsigned int
> a short (signed) to make things interesting
> a float because floating point is important
> a double as well
> another double but this time in big endian (network byte order)

获取数据 GET `/challenges/help_me_unpack/problem?access_token=aa8e14a982d0994e`

提交数据 POST `/challenges/help_me_unpack/solve?access_token=aa8e14a982d0994e`
