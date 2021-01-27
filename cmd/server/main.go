package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
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
)

func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development"
}

func getWpaSupplicant(ssid string, password string) string {
	replaced := strings.ReplaceAll(wpaSupplicant, "#ssid#", ssid)
	replaced = strings.ReplaceAll(replaced, "#password#", password)
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
		exec.Command(cmd)
	}

}

func main() {
	router := gin.Default()

	router.Static("/assets", "./cmd/server/assets")
	router.LoadHTMLGlob("./cmd/server/views/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	router.POST("/", func(c *gin.Context) {
		ssid := c.PostForm("ssid")
		password := c.PostForm("password")

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

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"ssid":     ssid,
			"password": password,
		})
	})

	router.Run()
}
