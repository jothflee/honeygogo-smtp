package core

import (
	"encoding/json"
	"net"
	"strings"
	"time"

	"github.com/oschwald/geoip2-golang"
	log "github.com/sirupsen/logrus"

	"github.com/emersion/go-smtp"
)

func StartSMTPServer(addr string) chan MessageMeta {
	outChannel := make(chan MessageMeta, 100)
	msgChannel := make(chan MessageMeta, 100)

	// maxCores = runtime.NumCPU()

	go func() {
		be := NewChannelBackend(msgChannel)

		s := smtp.NewServer(be)

		s.Addr = addr
		s.Domain = "localhost"
		s.ReadTimeout = 60 * time.Second
		s.WriteTimeout = 60 * time.Second
		s.MaxMessageBytes = 1024 * 1024
		s.MaxRecipients = 50
		s.AllowInsecureAuth = true

		log.Println("Starting server at", s.Addr)
		if err := s.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	// if you want to go big
	// for i := 0; i < maxCores * .25; i++ {
	for i := 0; i < 1; i++ {
		go func() {
			db, err := geoip2.Open("GeoLite2-City.mmdb")
			if err != nil {
				log.Fatal(err)
			}
			defer db.Close()

			for {
				select {
				case in := <-msgChannel:

					// If you are using strings that may be invalid, check that ip is not nil
					fromip := net.ParseIP(strings.Split(in.FromAddr, ":")[0])
					toip := net.ParseIP(strings.Split(in.FromAddr, ":")[0])
					record, err := db.City(fromip)
					if err != nil {
						log.Fatal(err)
					}
					in.Milis = time.Now().UTC().UnixNano() / int64(time.Millisecond)
					in.Location = GeoPoint{
						Latitude:  record.Location.Latitude,
						Longitude: record.Location.Longitude,
					}

					// decorated, move on with processing
					in.FromAddr = fromip.String()
					in.ToAddr = toip.String()
					outChannel <- in
				}
			}
		}()
	}

	return outChannel
}

func JSONstringify(in interface{}) string {
	return string(JSONBytify(in))
}
func JSONBytify(in interface{}) []byte {
	b, err := json.Marshal(in)
	if err != nil {
		b = []byte{}
	}
	return b
}
