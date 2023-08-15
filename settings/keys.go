package settings

import (
	"errors"
	"github.com/magiconair/properties"
	"path/filepath"
)

var (
	keysInstance *switchKeys
)

type switchKeys struct {
	keys map[string]string
}

func (k *switchKeys) GetKey(keyName string) string {
	return k.keys[keyName]
}

func SwitchKeys() (*switchKeys, error) {
	return keysInstance, nil
}

func InitSwitchKeys(dataFolder string) (*switchKeys, error) {
	settings := ReadSettings(dataFolder)
	path := settings.Prodkeys
	keys, err := GetSwitchKeys(path)

	if err != nil {
		return nil, errors.New("Error trying to read prod.keys [reason:" + err.Error() + "]")
	}

	settings.Prodkeys = path
	SaveSettings(settings, dataFolder)
	keysInstance = &switchKeys{keys: keys}

	return keysInstance, nil
}

func GetSwitchKeys(path string) (map[string]string, error) {
	keys := map[string]string{}

	p, err := properties.LoadFile(filepath.Join(path, "prod.keys"), properties.UTF8)

	if err != nil {
		return keys, err
	}

	for _, key := range p.Keys() {
		value, _ := p.Get(key)
		keys[key] = value
	}

	return keys, nil
}

func IsKeysFileAvailable() bool {
	if keys, _ := SwitchKeys(); keys != nil && keys.GetKey("header_key") != "" {
		return true
	}

	return false
}
