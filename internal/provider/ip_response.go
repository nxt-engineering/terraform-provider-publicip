package provider

type IPResponse struct {
	IP         string  `json:"ip,omitempty"`
	IPDecimal  int64   `json:"ip_decimal,omitempty"`
	Country    string  `json:"country,omitempty"`
	CountryISO string  `json:"country_iso,omitempty"`
	CountryEU  bool    `json:"country_eu,omitempty"`
	RegionName string  `json:"region_name,omitempty"`
	RegionCode string  `json:"region_code,omitempty"`
	ZIPCode    string  `json:"zip_code,omitempty"`
	City       string  `json:"city,omitempty"`
	Latitude   float32 `json:"latitude,omitempty"`
	Longitude  float32 `json:"longitude,omitempty"`
	TimeZone   string  `json:"time_zone,omitempty"`
	ASN        string  `json:"asn,omitempty"`
	ASNOrg     string  `json:"asn_org,omitempty"`
	UserAgent  struct {
		Product  string `json:"product,omitempty"`
		Version  string `json:"version,omitempty"`
		Comment  string `json:"comment,omitempty"`
		RAWValue string `json:"raw_value,omitempty"`
	} `json:"user_agent"`
}
