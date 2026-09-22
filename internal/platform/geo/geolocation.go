package geo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// LocationInfo is the normalized result of reverse geocoding coordinates.
type LocationInfo struct {
	Country     *string
	CountryCode *string
	Province    *string
	City        *string
	District    *string
	Address     *string
}

// GeocodingService resolves GPS coordinates to human-readable location fields.
type GeocodingService interface {
	ReverseGeocode(lat, lng float64) (*LocationInfo, error)
}

// AMapGeocodingService implements reverse geocoding through AMap.
type AMapGeocodingService struct {
	APIKey     string
	HTTPClient *http.Client
}

type amapResponse struct {
	Status    string `json:"status"`
	Info      string `json:"info"`
	Regeocode struct {
		AddressComponent struct {
			Country  string         `json:"country"`
			Province flexibleString `json:"province"`
			City     flexibleString `json:"city"`
			District flexibleString `json:"district"`
		} `json:"addressComponent"`
		FormattedAddress string `json:"formatted_address"`
	} `json:"regeocode"`
}

type flexibleString string

func (f *flexibleString) UnmarshalJSON(data []byte) error {
	var value string
	if json.Unmarshal(data, &value) == nil {
		*f = flexibleString(value)
		return nil
	}
	*f = ""
	return nil
}

func NewAMapGeocodingService(apiKey string) *AMapGeocodingService {
	return &AMapGeocodingService{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (s *AMapGeocodingService) ReverseGeocode(lat, lng float64) (*LocationInfo, error) {
	query := url.Values{
		"key":        {s.APIKey},
		"location":   {fmt.Sprintf("%f,%f", lng, lat)},
		"extensions": {"base"},
		"output":     {"json"},
	}
	response, err := s.HTTPClient.Get("https://restapi.amap.com/v3/geocode/regeo?" + query.Encode())
	if err != nil {
		return nil, fmt.Errorf("reverse geocode request failed: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read reverse geocode response failed: %w", err)
	}
	var result amapResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse reverse geocode response failed: %w", err)
	}
	if result.Status != "1" {
		return nil, fmt.Errorf("amap reverse geocode failed: %s", result.Info)
	}

	address := result.Regeocode.AddressComponent
	location := &LocationInfo{}
	if address.Country != "" {
		location.Country = &address.Country
		if address.Country == "中国" || address.Country == "China" {
			code := "CN"
			location.CountryCode = &code
		}
	}
	if address.Province != "" {
		value := string(address.Province)
		location.Province = &value
	}
	if address.City != "" {
		value := string(address.City)
		location.City = &value
	} else if address.Province != "" {
		value := string(address.Province)
		location.City = &value
	}
	if address.District != "" {
		value := string(address.District)
		location.District = &value
	}
	if result.Regeocode.FormattedAddress != "" {
		location.Address = &result.Regeocode.FormattedAddress
	}
	return location, nil
}
