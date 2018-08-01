package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/soracom/endorse-client-go/endorse"
)

var log = logging.MustGetLogger("endorse-cli")
var formatNormal = logging.MustStringFormatter(
	`%{color} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)
var formatDebug = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfile} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

type runMode int

const (
	runModeNormal runMode = iota
	runModeListCOMPorts
	runModeDeviceInfo
	runModeDoAuthentication
	runModeDoNothing
	runModeUnknown
)

type appConfig struct {
	Debug bool
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

func run() error {
	rm, appCfg, eCfg, err := parseFlags()
	if err != nil {
		return err
	}
	if rm == runModeDoNothing {
		return nil
	}

	setupLogger(appCfg)
	eCfg.Logger = log

	ec, err := endorse.NewClient(eCfg)
	if err != nil {
		return err
	}
	defer ec.Close()

	switch rm {
	case runModeListCOMPorts:
		return listCOMPorts(ec)
	case runModeDeviceInfo:
		return deviceInfo(ec)
	case runModeDoAuthentication:
		return doAuthentication(ec)
	default:
		return errors.New("unknown run mode")
	}
}

func parseFlags() (runMode, *appConfig, *endorse.Config, error) {
	var (
		keysAPIEndpointURL string
		uiccInterfaceType  string
		portName           string
		baudRate           uint
		dataBits           uint
		stopBits           uint
		parityMode         uint

		listCOMPorts bool
		deviceInfo   bool

		disableKeyCache bool
		clearKeyCache   bool

		help    bool
		version bool
		debug   bool
	)
	flag.StringVar(&keysAPIEndpointURL, "keys-api-endpoint-url", "", "Use the specified URL as a Keys API endpoint")
	flag.StringVar(&uiccInterfaceType, "interface", "autoDetect", "UICC Interface to use. Valid values are iso7816, comm, or autoDetect")
	flag.StringVar(&portName, "port-name", "", "Port name of communiation device (e.g. -c COM1 or -c /dev/tty1)")
	flag.UintVar(&baudRate, "baud-rate", 115200, "Baud rate for communiation device (e.g. -b 115200)")
	flag.UintVar(&dataBits, "data-bits", 8, "Data bits for communiation device (e.g. -s 8)")
	flag.UintVar(&stopBits, "stop-bits", 1, "Stop bits for communiation device (e.g. -s 1)")
	flag.UintVar(&parityMode, "parity-mode", 0, "Parity mode for communiation device. 0: None, 1: Odd, 2: Even")

	flag.BoolVar(&listCOMPorts, "list-com-ports", false, "List all available communication devices and exit")
	flag.BoolVar(&deviceInfo, "device-info", false, "Query the communication device and print the information")

	flag.BoolVar(&disableKeyCache, "disable-key-cache", false, "Do not store authentication result to the key cache")
	flag.BoolVar(&clearKeyCache, "clear-key-cache", false, "Remove all items in the key cache")

	flag.BoolVar(&help, "help", false, "Display this help message and exit")
	flag.BoolVar(&help, "h", false, "Display this help message and exit")
	flag.BoolVar(&version, "version", false, "Show version number")
	flag.BoolVar(&debug, "debug", false, "Show verbose debug messages")
	flag.Parse()

	if help {
		flag.Usage()
		return runModeDoNothing, nil, nil, nil
	}
	if version {
		showVersion()
		return runModeDoNothing, nil, nil, nil
	}

	var err error

	uit, err := endorse.ParseUICCInterfaceType(uiccInterfaceType)
	if err != nil {
		return runModeUnknown, nil, nil, err
	}
	if portName != "" {
		*uit = endorse.UICCInterfaceTypeComm
	}

	kc := endorse.KeyCacheConfig{
		Disabled: disableKeyCache,
		Clear:    clearKeyCache,
	}

	serial := endorse.SerialConfig{
		PortName:   portName,
		BaudRate:   baudRate,
		DataBits:   dataBits,
		StopBits:   stopBits,
		ParityMode: endorse.ParityMode(parityMode),
	}

	appCfg := &appConfig{
		Debug: debug,
	}

	var kaeu *url.URL
	if keysAPIEndpointURL != "" {
		kaeu, err = url.Parse(keysAPIEndpointURL)
		if err != nil {
			return runModeUnknown, nil, nil, err
		}
	}

	eCfg := &endorse.Config{
		KeysAPIEndpointURL: kaeu,
		UICCInterfaceType:  *uit,
		KeyCache:           kc,
		Serial:             serial,
		Logger:             log,
	}

	if listCOMPorts {
		eCfg.UICCInterfaceType = endorse.UICCInterfaceTypeNone
		return runModeListCOMPorts, appCfg, eCfg, nil
	}

	if deviceInfo {
		eCfg.UICCInterfaceType = endorse.UICCInterfaceTypeComm
		if portName == "" {
			return runModeUnknown, nil, nil, errors.New("-port-name must be specified with -device-info")
		}
		return runModeDeviceInfo, appCfg, eCfg, nil
	}

	return runModeDoAuthentication, appCfg, eCfg, nil
}

func setupLogger(appCfg *appConfig) {
	be := logging.NewLogBackend(os.Stderr, "", 0)
	format := formatNormal
	if appCfg.Debug {
		format = formatDebug
	}
	bf := logging.NewBackendFormatter(be, format)
	ml := logging.AddModuleLevel(bf)
	level := logging.ERROR
	if appCfg.Debug {
		level = logging.DEBUG
	}
	ml.SetLevel(level, "")
	logging.SetBackend(ml)
}

func listCOMPorts(ec *endorse.Client) error {
	ports, err := ec.ListCOMPorts()
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(ports, "\n"))
	return nil
}

func deviceInfo(ec *endorse.Client) error {
	di, err := ec.GetDeviceInfo()
	if err != nil {
		return err
	}

	fmt.Println(di)
	return nil
}

func doAuthentication(ec *endorse.Client) error {
	ar, err := ec.DoAuthentication()
	if err != nil {
		return err
	}

	b, err := json.Marshal(ar)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}

func showVersion() error {
	fmt.Println(Version)
	return nil
}
