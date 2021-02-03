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
	"time"

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
	rcLocal = `#!/bin/sh -e
#
# rc.local
#
# This script is executed at the end of each multiuser runlevel.
# Make sure that the script will "exit 0" on success or any other
# value on error.
#
# In order to enable or disable this script just change the execution
# bits.
#
# By default this script does nothing.

# Print the IP address
_IP=$(hostname -I) || true
if [ "$_IP" ]; then
  printf "My IP address is %s\n" "$_IP"
fi

# Start PICOM Client
/home/pi/bin/picom-client &

exit 0`
)

func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development"
}

func getPicomConfig(username string, server string, password string) string {
	replaced := strings.Replace(picomConfig, "#username#", username, -1)
	replaced = strings.Replace(replaced, "#server#", server, -1)
	replaced = strings.Replace(replaced, "#password#", password, -1)
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

func reboot() {
	time.Sleep(1 * time.Second)
	system("reboot")
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

		writeFile("/etc/rc.local", rcLocal, 0664)

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

		system("sudo systemctl disable dnsmasq")
		system("sudo systemctl disable hostapd")
		system("sudo systemctl disable picom-setup-server")

		system("sudo systemctl enable picom-client")

		go reboot()

		c.HTML(http.StatusOK, "index", gin.H{
			"reloading": true,
		})
	})

	router.Run()
}
