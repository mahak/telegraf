package nginx_plus_api

type processes struct {
	Respawned int `json:"respawned"`
}

type connections struct {
	Accepted int64 `json:"accepted"`
	Dropped  int64 `json:"dropped"`
	Active   int64 `json:"active"`
	Idle     int64 `json:"idle"`
}

type slabs map[string]struct {
	Pages struct {
		Used int64 `json:"used"`
		Free int64 `json:"free"`
	} `json:"pages"`
	Slots map[string]struct {
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
		Reqs  int64 `json:"reqs"`
		Fails int64 `json:"fails"`
	} `json:"slots"`
}

type ssl struct { // added in version 6
	Handshakes       int64 `json:"handshakes"`
	HandshakesFailed int64 `json:"handshakes_failed"`
	SessionReuses    int64 `json:"session_reuses"`
}

type resolverZones map[string]struct {
	Requests struct {
		Name int64 `json:"name"`
		Srv  int64 `json:"srv"`
		Addr int64 `json:"addr"`
	} `json:"requests"`
	Responses struct {
		Noerror  int64 `json:"noerror"`
		Formerr  int64 `json:"formerr"`
		Servfail int64 `json:"servfail"`
		Nxdomain int64 `json:"nxdomain"`
		Notimp   int64 `json:"notimp"`
		Refused  int64 `json:"refused"`
		Timedout int64 `json:"timedout"`
		Unknown  int64 `json:"unknown"`
	} `json:"responses"`
}

type httpRequests struct {
	Total   int64 `json:"total"`
	Current int64 `json:"current"`
}

type responseStats struct {
	Responses1xx int64 `json:"1xx"`
	Responses2xx int64 `json:"2xx"`
	Responses3xx int64 `json:"3xx"`
	Responses4xx int64 `json:"4xx"`
	Responses5xx int64 `json:"5xx"`
	Total        int64 `json:"total"`
}

type httpServerZones map[string]struct {
	Processing int           `json:"processing"`
	Requests   int64         `json:"requests"`
	Responses  responseStats `json:"responses"`
	Discarded  *int64        `json:"discarded"` // added in version 6
	Received   int64         `json:"received"`
	Sent       int64         `json:"sent"`
}

type httpLocationZones map[string]struct {
	Requests  int64         `json:"requests"`
	Responses responseStats `json:"responses"`
	Discarded *int64        `json:"discarded"` // added in version 6
	Received  int64         `json:"received"`
	Sent      int64         `json:"sent"`
}

type healthCheckStats struct {
	Checks     int64 `json:"checks"`
	Fails      int64 `json:"fails"`
	Unhealthy  int64 `json:"unhealthy"`
	LastPassed *bool `json:"last_passed"`
}

type httpUpstreams map[string]struct {
	Peers []struct {
		ID           *int             `json:"id"` // added in version 3
		Server       string           `json:"server"`
		Backup       bool             `json:"backup"`
		Weight       int              `json:"weight"`
		State        string           `json:"state"`
		Active       int              `json:"active"`
		Keepalive    *int             `json:"keepalive"` // removed in version 5
		MaxConns     *int             `json:"max_conns"` // added in version 3
		Requests     int64            `json:"requests"`
		Responses    responseStats    `json:"responses"`
		Sent         int64            `json:"sent"`
		Received     int64            `json:"received"`
		Fails        int64            `json:"fails"`
		Unavail      int64            `json:"unavail"`
		HealthChecks healthCheckStats `json:"health_checks"`
		Downtime     int64            `json:"downtime"`
		HeaderTime   *int64           `json:"header_time"`   // added in version 5
		ResponseTime *int64           `json:"response_time"` // added in version 5
	} `json:"peers"`
	Keepalive int       `json:"keepalive"`
	Zombies   int       `json:"zombies"` // added in version 6
	Queue     *struct { // added in version 6
		Size      int   `json:"size"`
		MaxSize   int   `json:"max_size"`
		Overflows int64 `json:"overflows"`
	} `json:"queue"`
}

type streamServerZones map[string]struct {
	Processing  int            `json:"processing"`
	Connections int            `json:"connections"`
	Sessions    *responseStats `json:"sessions"`
	Discarded   *int64         `json:"discarded"` // added in version 7
	Received    int64          `json:"received"`
	Sent        int64          `json:"sent"`
}

type streamUpstreams map[string]struct {
	Peers []struct {
		ID            int              `json:"id"`
		Server        string           `json:"server"`
		Backup        bool             `json:"backup"`
		Weight        int              `json:"weight"`
		State         string           `json:"state"`
		Active        int              `json:"active"`
		Connections   int64            `json:"connections"`
		ConnectTime   *int             `json:"connect_time"`
		FirstByteTime *int             `json:"first_byte_time"`
		ResponseTime  *int             `json:"response_time"`
		Sent          int64            `json:"sent"`
		Received      int64            `json:"received"`
		Fails         int64            `json:"fails"`
		Unavail       int64            `json:"unavail"`
		HealthChecks  healthCheckStats `json:"health_checks"`
		Downtime      int64            `json:"downtime"`
	} `json:"peers"`
	Zombies int `json:"zombies"`
}

type basicHitStats struct {
	Responses int64 `json:"responses"`
	Bytes     int64 `json:"bytes"`
}

type extendedHitStats struct {
	basicHitStats
	ResponsesWritten int64 `json:"responses_written"`
	BytesWritten     int64 `json:"bytes_written"`
}

type httpCaches map[string]struct { // added in version 2
	Size        int64            `json:"size"`
	MaxSize     int64            `json:"max_size"`
	Cold        bool             `json:"cold"`
	Hit         basicHitStats    `json:"hit"`
	Stale       basicHitStats    `json:"stale"`
	Updating    basicHitStats    `json:"updating"`
	Revalidated *basicHitStats   `json:"revalidated"` // added in version 3
	Miss        extendedHitStats `json:"miss"`
	Expired     extendedHitStats `json:"expired"`
	Bypass      extendedHitStats `json:"bypass"`
}

type httpLimitReqs map[string]struct {
	Passed         int64 `json:"passed"`
	Delayed        int64 `json:"delayed"`
	Rejected       int64 `json:"rejected"`
	DelayedDryRun  int64 `json:"delayed_dry_run"`
	RejectedDryRun int64 `json:"rejected_dry_run"`
}
