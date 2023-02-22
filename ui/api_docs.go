package ui

type SwaggerConfig struct {
	ConfigUrl              string        `json:"configUrl"`
	DisplayRequestDuration bool          `json:"displayRequestDuration"`
	Oauth2RedirectUrl      string        `json:"oauth2RedirectUrl"`
	OperationsSorter       string        `json:"operationsSorter"`
	ValidatorUrl           string        `json:"validatorUrl"`
	Urls                   *[]SwaggerUrl `json:"urls"`
}

type SwaggerUrl struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}
