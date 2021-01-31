package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/gobuffalo/packr"
)

const (
	wpaSupplicant = `ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1
country=AT

network={
	ssid="#ssid#"
	psk="#password#"
}`
	dhcpcdConfClient = `hostname
clientid
persistent
option rapid_commit
option domain_name_servers, domain_name, domain_search, host_name
option classless_static_routes
option ntp_servers
option interface_mtu
require dhcp_server_identifier
slaac private`
	picomConfig = `#server#
#username#
#password#
`
)

func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development"
}

func getPicomConfig(username string, server string, server_password string) string {
	replaced := strings.Replace(picomConfig, "#username#", username, -1)
	replaced = strings.Replace(replaced, "#server#", server, -1)
	replaced = strings.Replace(replaced, "#server_password#", server_password, -1)
	return replaced
}

func getWpaSupplicant(ssid string, password string) string {
	replaced := strings.Replace(wpaSupplicant, "#ssid#", ssid, -1)
	replaced = strings.Replace(replaced, "#password#", password, -1)
	return replaced
}

func writeFile(filename string, data string, perm os.FileMode) {
	if isDevelopment() {
		log.Println("Write file: " + filename)
	} else {
		ioutil.WriteFile(filename, []byte(data), perm)
	}
}

func system(cmd string) {
	if isDevelopment() {
		log.Println("System: " + cmd)
	} else {
		exec.Command(cmd).Run()
	}

}

func main() {
	router := gin.Default()

	usr, _ := user.Current()

	configSavePath := flag.String("saveConfigPath", usr.HomeDir+"/.picom", "path to save generated config")
	flag.Parse()

	views := packr.NewBox("./views")
	s, _ := views.FindString("index.tmpl")
	html := template.Must(template.New("index").Parse(s))

	router.SetHTMLTemplate(html)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index", gin.H{
			"title":     "Main website",
			"reloading": false,
		})
	})

	router.POST("/", func(c *gin.Context) {
		ssid := c.PostForm("ssid")
		password := c.PostForm("password")

		username := c.PostForm("username")
		server := c.PostForm("server")
		server_password := c.PostForm("server_password")

		writeFile(*configSavePath, getPicomConfig(username, server, server_password), 0644)

		writeFile("/etc/wpa_supplicant/wpa_supplicant.conf", getWpaSupplicant(ssid, password), 0644)
		system("chown root.root /etc/wpa_supplicant/wpa_supplicant.conf")
		system("chmod 600 /etc/wpa_supplicant/wpa_supplicant.conf")

		writeFile("/etc/dhcpcd.conf", dhcpcdConfClient, 0644)
		system("chmod 600 /etc/dhcpcd.conf")
		system("chown root.netdev /etc/dhcpcd.conf")

		system("chown root.root /etc/hostapd/hostapd.conf")
		system("chmod 644 /etc/hostapd/hostapd.conf")

		system("chown root.root /etc/dnsmasq.conf")
		system("chmod 644 /etc/dnsmasq.conf")

		system("chown root.root /etc/default/hostapd")
		system("chmod 644 /etc/default/hostapd")

		system("systemctl disable dnsmasq")
		system("systemctl disable hostapd")

		system("reboot")

		c.HTML(http.StatusOK, "index", gin.H{
			"ssid":      ssid,
			"password":  password,
			"reloading": true,
		})
	})

	router.Run()
}
