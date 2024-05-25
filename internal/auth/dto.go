package auth

type AgentSendSMSCodeRequest struct {
	Phone string `json:"phone" xml:"phone" form:"phone"`
}

type AgentCheckSMSCodeRequest struct {
	SMSCode string `json:"sms_code" xml:"sms_code" form:"sms_code"`
	Phone   string `json:"phone" xml:"phone" form:"phone"`
}

type AdminSetPasswordRequest struct {
	Phone    string `json:"phone" xml:"phone" form:"phone"`
	SMSCode  string `json:"sms_code" xml:"sms_code" form:"sms_code"`
	Password string `json:"password" xml:"password" form:"password"`
}
