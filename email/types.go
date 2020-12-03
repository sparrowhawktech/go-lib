package email

type MessageData struct {
	Recipient   *string `json:"recipient"`
	Subject     *string `json:"subject"`
	ContentType *string `json:"contentType"`
	Body        []byte  `json:"body"`
}
