package mail

type SES struct{}

func (s *SES) Send(email Email) error {
	return nil
}
