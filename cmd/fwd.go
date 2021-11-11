package cmd

// TODO: interface it
type Forwarder struct {
	Address  string
	Host     string
	User     string
	Password string
	TLS      bool
}

func parseFwder() (fwd *Forwarder) {
	if Fwd != "" && FwdServer != "" {
		fwd = &Forwarder{
			Address:  Fwd,
			Host:     FwdServer,
			User:     FwdUser,
			Password: FwdPw,
			TLS:      FwdTLS,
		}
	}
	return fwd
}
