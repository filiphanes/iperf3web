package main

// PublicServer is a known public iperf3 server.
type PublicServer struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Location string `json:"location"`
}

var publicServers = []PublicServer{
	{Name: "Bouygues Telecom", Host: "bouygues.iperf.fr", Port: 5201, Location: "Paris, FR"},
	{Name: "Online.net / Scaleway", Host: "ping.online.net", Port: 5200, Location: "Paris, FR"},
	{Name: "Moji.fr", Host: "iperf3.moji.fr", Port: 5209, Location: "Strasbourg, FR"},
	{Name: "Hurricane Electric", Host: "iperf.he.net", Port: 5201, Location: "Fremont, CA, US"},
	{Name: "Clouvider New York", Host: "nyc.speedtest.clouvider.net", Port: 5200, Location: "New York, US"},
	{Name: "Clouvider Los Angeles", Host: "la.speedtest.clouvider.net", Port: 5200, Location: "Los Angeles, US"},
	{Name: "Clouvider London", Host: "lon.speedtest.clouvider.net", Port: 5200, Location: "London, UK"},
	{Name: "Clouvider Amsterdam", Host: "ams.speedtest.clouvider.net", Port: 5200, Location: "Amsterdam, NL"},
	{Name: "Leaseweb Amsterdam", Host: "iperf.leaseweb.net", Port: 5201, Location: "Amsterdam, NL"},
	{Name: "wtnet.de", Host: "speedtest.wtnet.de", Port: 5200, Location: "Hamburg, DE"},
	{Name: "Serverius", Host: "speedtest.serverius.net", Port: 5002, Location: "Netherlands"},
	{Name: "IT-North Gothenburg", Host: "iperf.it-north.net", Port: 5201, Location: "Gothenburg, SE"},
	{Name: "i3D.net Rotterdam", Host: "speedtest.i3d.net", Port: 5201, Location: "Rotterdam, NL"},
}
