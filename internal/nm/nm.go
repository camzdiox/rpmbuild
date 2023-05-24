package nm

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/nm/connection"
)

// previousCon gets a connection object using the previous UUID
func previousCon(pUUID string) (*connection.Connection, error) {
	if pUUID == "" {
		return nil, errors.New("UUID is empty")
	}
	s, err := connection.NewSettings()
	if err != nil {
		return nil, err
	}
	return s.ConnectionByUUID(pUUID)
}

// createCon creates a new connection using the arguments
// if a previous connection was found with pUUID, it updates that one instead
// it returns the newly created or updated connection object
func createCon(pUUID string, args connection.SettingsArgs) (*connection.Connection, error) {
	prev, err := previousCon(pUUID)
	// previous connection found, update it with the new settings args
	if err == nil {
		return prev, prev.Update(args)
	}
	// create a connection settings object
	s, err := connection.NewSettings()
	if err != nil {
		return nil, err
	}
	// create a new connection
	return s.AddConnection(args)
}

// Install installs a non TLS network and returns an error if it cannot configure it
// Right now it adds a new profile that is not automatically added
// It returns the uuid if the connection was added successfully
func Install(n network.NonTLS, pUUID string) (string, error) {
	fID := fmt.Sprintf("%s (from Geteduroam)", n.SSID)
	cUser, err := user.Current()
	if err != nil {
		return "", err
	}
	sCon := map[string]interface{}{
		"permissions": []string{
			fmt.Sprintf("user:%s", cUser.Username),
		},
		"type": "802-11-wireless",
		"id":   fID,
	}
	sWifi := map[string]interface{}{
		"ssid":     []byte(n.SSID),
		"security": "802-11-wireless-security",
	}
	sWsec := map[string]interface{}{
		"key-mgmt": "wpa-eap",
		"proto":    []string{"rsn"},
		"pairwise": []string{strings.ToLower(n.MinRSN)},
		"group":    []string{strings.ToLower(n.MinRSN)},
	}
	var sids []string

	for _, sid := range n.ServerIDs {
		v := fmt.Sprintf("DNS:%s", sid)
		sids = append(sids, v)
	}
	s8021x := map[string]interface{}{
		"eap": []string{
			n.Method().String(),
		},
		"identity":           n.Credentials.Username,
		"ca-cert":            n.Certs.ToPEM(),
		"anonymous-identity": n.AnonIdentity,
		"password":           n.Credentials.Password,
		"password-flags":     0,
		"altsubject-matches": sids,
	}
	if n.InnerAuth.EAP() && n.MethodType == method.TTLS {
		s8021x["phase2-autheap"] = n.InnerAuth.String()
	} else {
		s8021x["phase2-auth"] = n.InnerAuth.String()
	}
	sIP4 := map[string]interface{}{
		"method": "auto",
	}
	sIP6 := map[string]interface{}{
		"method": "auto",
	}
	settings := map[string]map[string]interface{}{
		"connection":               sCon,
		"802-11-wireless":          sWifi,
		"802-11-wireless-security": sWsec,
		"802-1x":                   s8021x,
		"ipv4":                     sIP4,
		"ipv6":                     sIP6,
	}
	con, err := createCon(pUUID, settings)
	if err != nil {
		return "", err
	}
	// get the settings from the added connection
	gs, err := con.GetSettings()
	if err != nil {
		return "", err
	}
	uuid, err := gs.UUID()
	if err != nil {
		return "", err
	}
	return uuid, nil
}
