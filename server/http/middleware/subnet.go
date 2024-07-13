package middleware

import (
	"net"
	"net/http"
)

func Subnet(ipSubnet net.IP, next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if ipSubnet == nil {
			next.ServeHTTP(rw, req)

			return
		}

		ipXRealStr := req.Header.Get("X-Real-IP")
		ipXReal := net.ParseIP(ipXRealStr)

		ipRemoteStr, _, err := net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			http.Error(rw, "split remote IP", http.StatusInternalServerError)
		}

		ipRemote := net.ParseIP(ipRemoteStr)

		if !net.IP.Equal(ipXReal, ipSubnet) || !net.IP.Equal(ipXReal, ipRemote) {
			http.Error(rw, "not trusted subnet", http.StatusForbidden)

			return
		}

		next.ServeHTTP(rw, req)
	}
}
