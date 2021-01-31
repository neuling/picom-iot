package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const (
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

# Start PICOM Setup Server
sudo /home/pi/bin/picom-iot-server &

exit 0`
)

func isDevelopment() bool {
	env := os.Getenv("ENV")
	return env == "development"
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
	system("rm -f /etc/wpa_supplicant/wpa_supplicant.conf")

	writeFile("/etc/dhcpcd.conf", dhcpcdConfHost, 0664)
	writeFile("/etc/hostapd/hostapd.conf", hostapdConf, 0664)
	writeFile("/etc/dnsmasq.conf", dnsmasqConf, 0664)
	writeFile("/etc/default/hostapd", defaultHostapd, 0664)

	writeFile("/etc/rc.local", rcLocal, 0664)

	system("chown root.netdev /etc/dhcpcd.conf")
	system("chown root.root /etc/hostapd/hostapd.conf")
	system("chown root.root /etc/dnsmasq.conf")
	system("chown root.root /etc/default/hostapd")

	system("systemctl enable dnsmasq")
	system("systemctl unmask hostapd.service")
	system("systemctl enable hostapd")

	system("reboot")
}
