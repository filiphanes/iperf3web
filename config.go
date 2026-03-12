package main

// PublicServer is a known public iperf3 server.
type PublicServer struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Location string `json:"location"`
}

var publicServers = []PublicServer{
	// Europe
	{Host: "bouygues.iperf.fr", Port: 5201, Location: "Paris, FR"},
	{Host: "ping.online.net", Port: 5200, Location: "Paris, FR"},
	{Host: "iperf3.moji.fr", Port: 5209, Location: "Strasbourg, FR"},
	{Host: "speedtest.init7.net", Port: 5201, Location: "Winterthur, CH"},
	{Host: "fra.speedtest.clouvider.net", Port: 5200, Location: "Frankfurt, DE"},
	{Host: "speedtest.wtnet.de", Port: 5200, Location: "Hamburg, DE"},
	{Host: "lon.speedtest.clouvider.net", Port: 5200, Location: "London, UK"},
	{Host: "ams.speedtest.clouvider.net", Port: 5200, Location: "Amsterdam, NL"},
	{Host: "speedtest.ams1.nl.leaseweb.net", Port: 5201, Location: "Amsterdam, NL"},
	{Host: "speedtest.serverius.net", Port: 5002, Location: "Netherlands"},
	{Host: "speedtest.i3d.net", Port: 5201, Location: "Rotterdam, NL"},
	{Host: "iperf.it-north.net", Port: 5201, Location: "Gothenburg, SE"},
	{Host: "iperf.volia.net", Port: 5201, Location: "Kyiv, UA"},
	// North America
	{Host: "iperf.he.net", Port: 5201, Location: "Fremont, CA, US"},
	{Host: "la.speedtest.clouvider.net", Port: 5200, Location: "Los Angeles, US"},
	{Host: "nyc.speedtest.clouvider.net", Port: 5200, Location: "New York, US"},
	{Host: "ash.speedtest.clouvider.net", Port: 5200, Location: "Ashburn, US"},
	{Host: "speedtest.chi11.us.leaseweb.net", Port: 5201, Location: "Chicago, US"},
	{Host: "speedtest.nocix.net", Port: 5201, Location: "Kansas City, US"},
	{Host: "speedtest.mtl2.ca.leaseweb.net", Port: 5201, Location: "Montreal, CA"},
	// Asia-Pacific
	{Host: "speedtest.uztelecom.uz", Port: 5200, Location: "Tashkent, UZ"},
	{Host: "speedtest.hkg12.hk.leaseweb.net", Port: 5201, Location: "Hong Kong, HK"},
	{Host: "speedtest.tyo11.jp.leaseweb.net", Port: 5201, Location: "Tokyo, JP"},
	{Host: "speedtest.sin1.sg.leaseweb.net", Port: 5201, Location: "Singapore, SG"},
	{Host: "speedtest.syd12.au.leaseweb.net", Port: 5201, Location: "Sydney, AU"},
}
