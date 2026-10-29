package types

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/PretendoNetwork/nex-go/v2/constants"
)




type StationURL struct {
	urlType        constants.StationURLType
	url            string
	flags          uint8
	standardParams map[string]string
	customParams   map[string]string
}

func (s *StationURL) ensureFields() {
	if s.standardParams == nil {
		s.standardParams = make(map[string]string)
	}

	if s.customParams == nil {
		s.customParams = make(map[string]string)
	}
}

func (s StationURL) numberParamValue(name string, bits int) (uint64, bool) {
	valueString, ok := s.ParamValue(name)
	if !ok {
		return 0, false
	}

	value, err := strconv.ParseUint(valueString, 10, bits)
	if err != nil {
		return 0, false
	}

	return value, true
}

func (s StationURL) uint8ParamValue(name string) (uint8, bool) {
	value, ok := s.numberParamValue(name, 8)
	if !ok {
		return 0, false
	}

	return uint8(value), true
}

func (s StationURL) uint16ParamValue(name string) (uint16, bool) {
	value, ok := s.numberParamValue(name, 16)
	if !ok {
		return 0, false
	}

	return uint16(value), true
}

func (s StationURL) uint32ParamValue(name string) (uint32, bool) {
	value, ok := s.numberParamValue(name, 32)
	if !ok {
		return 0, false
	}

	return uint32(value), true
}

func (s StationURL) uint64ParamValue(name string) (uint64, bool) {
	return s.numberParamValue(name, 64)
}

func (s StationURL) boolParamValue(name string) bool {
	valueString, ok := s.ParamValue(name)
	if !ok {
		return false
	}

	return valueString == "1"
}


func (s StationURL) WriteTo(writable Writable) {
	url := NewString(s.URL())

	url.WriteTo(writable)
}


func (s *StationURL) ExtractFrom(readable Readable) error {
	s.ensureFields()

	url := NewString("")

	if err := url.ExtractFrom(readable); err != nil {
		return fmt.Errorf("Failed to read StationURL. %s", err.Error())
	}

	s.SetURL(string(url))
	s.Parse()

	return nil
}


func (s StationURL) Copy() RVType {
	return NewStationURL(String(s.URL()))
}


func (s StationURL) Equals(o RVType) bool {
	if _, ok := o.(StationURL); !ok {
		return false
	}

	other := o.(StationURL)

	if s.urlType != other.urlType {
		return false
	}

	if s.flags != other.flags {
		return false
	}

	if len(s.standardParams) != len(other.standardParams) {
		return false
	}

	for key, value1 := range s.standardParams {
		value2, ok := other.standardParams[key]
		if !ok || value1 != value2 {
			return false
		}
	}

	return true
}



func (s StationURL) CopyRef() RVTypePtr {
	copied := s.Copy().(StationURL)
	return &copied
}




func (s *StationURL) Deref() RVType {
	return *s
}





func (s *StationURL) Set(name, value string, custom bool) {
	if custom {
		s.customParams[name] = value
	} else {
		s.standardParams[name] = value
	}
}







func (s *StationURL) Get(name string, custom bool) (string, bool) {
	var m map[string]string

	if custom {
		m = s.customParams
	} else {
		m = s.standardParams
	}

	if value, ok := m[name]; ok {
		return value, true
	}

	return "", false
}


func (s *StationURL) SetParamValue(name, value string) {
	s.standardParams[name] = value
}




func (s *StationURL) RemoveParam(name string) {
	delete(s.standardParams, name)
}






func (s StationURL) ParamValue(name string) (string, bool) {
	if value, ok := s.standardParams[name]; ok {
		return value, true
	}

	return "", false
}


func (s *StationURL) SetAddress(address string) {
	s.SetParamValue("address", address)
}




func (s StationURL) Address() (string, bool) {
	return s.ParamValue("address")
}


func (s *StationURL) SetPortNumber(port uint16) {
	s.SetParamValue("port", strconv.FormatUint(uint64(port), 10))
}






func (s *StationURL) PortNumber() (uint16, bool) {
	return s.uint16ParamValue("port")
}


func (s *StationURL) SetURLType(urlType constants.StationURLType) {
	s.urlType = urlType
}




func (s StationURL) URLType() constants.StationURLType {
	return s.urlType
}




func (s *StationURL) SetStreamID(streamID uint8) {
	s.SetParamValue("sid", strconv.FormatUint(uint64(streamID), 10))
}








func (s StationURL) StreamID() (uint8, bool) {
	return s.uint8ParamValue("sid")
}




func (s *StationURL) SetStreamType(streamType constants.StreamType) {
	s.SetParamValue("stream", strconv.FormatUint(uint64(streamType), 10))
}








func (s StationURL) StreamType() (constants.StreamType, bool) {
	streamType, ok := s.uint8ParamValue("stream")



	return constants.StreamType(streamType), ok
}




func (s *StationURL) SetNodeID(nodeID uint16) {
	s.SetParamValue("NodeID", strconv.FormatUint(uint64(nodeID), 10))
}






func (s StationURL) NodeID() (uint16, bool) {
	return s.uint16ParamValue("NodeID")
}


func (s *StationURL) SetPrincipalID(pid PID) {
	s.SetParamValue("PID", strconv.FormatUint(uint64(pid), 10))
}






func (s StationURL) PrincipalID() (PID, bool) {
	pid, ok := s.uint64ParamValue("PID")
	if !ok {
		return NewPID(0), false
	}

	return NewPID(pid), true
}




func (s *StationURL) SetConnectionID(connectionID uint32) {
	s.SetParamValue("CID", strconv.FormatUint(uint64(connectionID), 10))
}






func (s StationURL) ConnectionID() (uint32, bool) {
	return s.uint32ParamValue("CID")
}




func (s *StationURL) SetRVConnectionID(connectionID uint32) {
	s.SetParamValue("RVCID", strconv.FormatUint(uint64(connectionID), 10))
}






func (s StationURL) RVConnectionID() (uint32, bool) {
	return s.uint32ParamValue("RVCID")
}


func (s *StationURL) SetProbeRequestID(probeRequestID uint32) {
	s.SetParamValue("PRID", strconv.FormatUint(uint64(probeRequestID), 10))
}






func (s StationURL) ProbeRequestID() (uint32, bool) {
	return s.uint32ParamValue("PRID")
}


func (s *StationURL) SetFastProbeResponse(fast bool) {
	if fast {
		s.SetParamValue("fastproberesponse", "1")
	} else {
		s.SetParamValue("fastproberesponse", "0")
	}
}




func (s StationURL) IsFastProbeResponseEnabled() bool {
	return s.boolParamValue("fastproberesponse")
}


func (s *StationURL) SetNATMapping(mapping constants.NATMappingProperties) {
	s.SetParamValue("natm", strconv.FormatUint(uint64(mapping), 10))
}






func (s StationURL) NATMapping() (constants.NATMappingProperties, bool) {
	natm, ok := s.uint8ParamValue("natm")



	return constants.NATMappingProperties(natm), ok
}


func (s *StationURL) SetNATFiltering(filtering constants.NATFilteringProperties) {
	s.SetParamValue("natf", strconv.FormatUint(uint64(filtering), 10))
}






func (s StationURL) NATFiltering() (constants.NATFilteringProperties, bool) {
	natf, ok := s.uint8ParamValue("natf")



	return constants.NATFilteringProperties(natf), ok
}


func (s *StationURL) SetProbeRequestInitiation(probeinit bool) {
	if probeinit {
		s.SetParamValue("probeinit", "1")
	} else {
		s.SetParamValue("probeinit", "0")
	}
}




func (s StationURL) IsProbeRequestInitiationEnabled() bool {
	return s.boolParamValue("probeinit")
}


func (s *StationURL) SetUPnPSupport(supported bool) {
	if supported {
		s.SetParamValue("upnp", "1")
	} else {
		s.SetParamValue("upnp", "0")
	}
}




func (s StationURL) IsUPnPSupported() bool {
	return s.boolParamValue("upnp")
}




func (s *StationURL) SetNATPMPSupport(supported bool) {
	if supported {
		s.SetParamValue("pmp", "1")
	} else {
		s.SetParamValue("pmp", "0")
	}
}




func (s StationURL) IsNATPMPSupported() bool {
	return s.boolParamValue("pmp")
}


func (s *StationURL) SetURL(url string) {
	s.url = url
}




func (s StationURL) URL() string {
	s.Format()
	return s.url
}


func (s *StationURL) SetType(flags uint8) {
	s.flags = flags
	s.SetParamValue("type", strconv.FormatUint(uint64(flags), 10))
}






func (s StationURL) Type() (uint8, bool) {
	return s.uint8ParamValue("type")
}


func (s *StationURL) SetRelayServerAddress(address string) {
	s.SetParamValue("Rsa", address)
}




func (s StationURL) RelayServerAddress() (string, bool) {
	return s.ParamValue("Rsa")
}


func (s *StationURL) SetRelayServerPort(port uint16) {
	s.SetParamValue("Rsp", strconv.FormatUint(uint64(port), 10))
}






func (s StationURL) RelayServerPort() (uint16, bool) {
	return s.uint16ParamValue("Rsp")
}


func (s *StationURL) SetRelayAddress(address string) {
	s.SetParamValue("Ra", address)
}




func (s StationURL) RelayAddress() (string, bool) {
	return s.ParamValue("Ra")
}


func (s *StationURL) SetRelayPort(port uint16) {
	s.SetParamValue("Rp", strconv.FormatUint(uint64(port), 10))
}






func (s StationURL) RelayPort() (uint16, bool) {
	return s.uint16ParamValue("Rp")
}


func (s *StationURL) SetUseRelayServer(useRelayServer bool) {
	if useRelayServer {
		s.SetParamValue("R", "1")
	} else {
		s.SetParamValue("R", "0")
	}
}




func (s StationURL) IsRelayServerEnabled() bool {
	return s.boolParamValue("R")
}


func (s *StationURL) SetPlatformType(platformType uint8) {



	s.SetParamValue("Pl", strconv.FormatUint(uint64(platformType), 10))
}






func (s StationURL) PlatformType() (uint8, bool) {
	return s.uint8ParamValue("Pl")
}


func (s StationURL) IsPublic() bool {
	return s.flags&uint8(constants.StationURLFlagPublic) == uint8(constants.StationURLFlagPublic)
}


func (s StationURL) IsBehindNAT() bool {
	return s.flags&uint8(constants.StationURLFlagBehindNAT) == uint8(constants.StationURLFlagBehindNAT)
}


func (s *StationURL) Parse() {
	url := s.url
	if url == "" || len(url) > 1024 {

		return
	}

	parts := strings.SplitN(string(url), ":/", 2)



	if len(parts) != 2 {
		return
	}

	scheme := parts[0]
	parametersString := parts[1]

	switch scheme {
	case "prudp":
		s.SetURLType(constants.StationURLPRUDP)
	case "prudps":
		s.SetURLType(constants.StationURLPRUDPS)
	case "udp":
		s.SetURLType(constants.StationURLUDP)
	default:
		return
	}


	if parametersString == "" {
		return
	}

	parts = strings.SplitN(parametersString, "#", 2)
	standardSection := parts[0]
	customSection := ""

	if len(parts) == 2 {
		customSection = parts[1]
	}

	standardParameters := strings.Split(standardSection, ";")

	for i := range standardParameters {
		key, value, _ := strings.Cut(standardParameters[i], "=")

		if key == "address" && len(value) > 256 {


			return
		}

		if key == "port" {
			if port, err := strconv.Atoi(value); err != nil || (port < 0 || port > 65535) {

				return
			}
		}

		s.Set(key, value, false)
	}

	if len(customSection) != 0 {
		customParameters := strings.Split(customSection, ";")

		for i := range customParameters {
			key, value, _ := strings.Cut(customParameters[i], "=")

			s.Set(key, value, true)
		}
	}

	if flags, ok := s.uint8ParamValue("type"); ok {
		s.flags = flags
	}
}


func (s *StationURL) Format() {
	scheme := ""



	if s.urlType == constants.StationURLPRUDP {
		scheme = "prudp:/"
	} else if s.urlType == constants.StationURLPRUDPS {
		scheme = "prudps:/"
	} else if s.urlType == constants.StationURLUDP {
		scheme = "udp:/"
	}

	fields := make([]string, 0)

	for key, value := range s.standardParams {
		if key == "address" && len(value) > 256 {


			return
		}

		if key == "port" {
			if port, err := strconv.Atoi(value); err != nil || (port < 0 || port > 65535) {

				return
			}
		}

		fields = append(fields, fmt.Sprintf("%s=%s", key, value))
	}

	url := scheme + strings.Join(fields, ";")

	if len(s.customParams) != 0 {
		customFields := make([]string, 0)

		for key, value := range s.customParams {
			customFields = append(customFields, fmt.Sprintf("%s=%s", key, value))
		}

		url = url + "#" + strings.Join(customFields, ";")
	}

	if len(url) > 1024 {

		return
	}

	s.url = url
}


func (s StationURL) String() string {
	return s.FormatToString(0)
}


func (s StationURL) FormatToString(indentationLevel int) string {
	indentationValues := strings.Repeat("\t", indentationLevel+1)
	indentationEnd := strings.Repeat("\t", indentationLevel)

	var b strings.Builder

	b.WriteString("StationURL{\n")
	b.WriteString(fmt.Sprintf("%surl: %q\n", indentationValues, s.URL()))
	b.WriteString(fmt.Sprintf("%s}", indentationEnd))

	return b.String()
}






func (s StationURL) Value() (driver.Value, error) {
	return s.URL(), nil
}




func (s *StationURL) Scan(value any) error {
	s.ensureFields()

	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		s.SetURL(v)
	case []byte:
		s.SetURL(string(v))
	default:
		return fmt.Errorf("cannot scan type %T into StationURL", v)
	}

	s.Parse()

	return nil
}


func NewStationURL(url String) StationURL {
	stationURL := StationURL{
		url:            string(url),
		standardParams: make(map[string]string),
		customParams:   make(map[string]string),
	}

	stationURL.Parse()

	return stationURL
}
