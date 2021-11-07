package backend

import "github.com/jothflee/honeygogo/core"

type Backend interface {
	OnMessage(core.MessageMeta)
	Close()
}
