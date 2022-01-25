provider "publicip" {
  provider_url = "https://ifconfig.co/" # optional
  timeout      = "10s"                  # optional

  # 1 request per 500ms
  rate_limit_rate  = "500ms" # optional
  rate_limit_burst = "1"     # optional
}
