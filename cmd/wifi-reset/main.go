package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	wpaSupplicant = `ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1

network={
	ssid="#ssid#"
	psk="#password#"
}`
	defaultHostapd = `DAEMON_CONF="/etc/hostapd/hostapd.conf`
	dnsmasqConf    = `interface=wlan0
dhcp-range=10.0.0.2,10.0.0.20,255.255.255.0,24h`
	hostapdConf = `interface=wlan0
driver=nl80211
ssid=picom
hw_mode=g
channel=7
wmm_enabled=0
macaddr_acl=0
auth_algs=1
ignore_broadcast_ssid=0`
	dhcpcdConfHost = `hostname
clientid
persistent
option rapid_commit
option domain_name_servers, domain_name, domain_search, host_name
option classless_static_routes
option ntp_servers
option interface_mtu
require dhcp_server_identifier
slaac private
interface wlan0
static ip_address=10.0.0.1/24`
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
	system("rm -f /etc/wpa_supplicant/wpa_supplicant.conf")

	writeFile("/etc/dhcpcd.conf", dhcpcdConfHost, 0664)
	writeFile("/etc/hostapd/hostapd.conf", defaultHostapd, 0664)
	writeFile("/etc/dnsmasq.conf", dnsmasqConf, 0664)
	writeFile("/etc/default/hostapd", defaultHostapd, 0664)

	system("chown root.netdev /etc/dhcpcd.conf")
	system("chown root.root /etc/hostapd/hostapd.conf")
	system("chown root.root /etc/dnsmasq.conf")
	system("chown root.root /etc/default/hostapd")

	system("systemctl enable dnsmasq")
	system("systemctl enable hostapd")

	system("reboot")
}
