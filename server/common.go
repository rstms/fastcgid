// go-common local proxy functions

package server

import (
	rstms "github.com/rstms/go-common"
)

type APIClient interface {
	Close()
	Get(path string, response interface{}) (string, error)
	Post(path string, request, response interface{}, headers *map[string]string) (string, error)
	Put(path string, request, response interface{}, headers *map[string]string) (string, error)
	Delete(path string, response interface{}) (string, error)
	SetFlag(string, bool) error
	StatusCode() (int, bool)
}

type CobraCommand interface {
}

type Sendmail interface {
	Send(to, from, subject string, body []byte) error
}

func NewAPIClient(prefix, url, certFile, keyFile, caFile string, headers *map[string]string) (APIClient, error) {
	return rstms.NewAPIClient(prefix, url, certFile, keyFile, caFile, headers)
}

func OptionKey(cobraCmd CobraCommand, key string) string {
	return rstms.OptionKey(cobraCmd, key)
}

func OptionSwitch(cobraCmd CobraCommand, name, flag, description string) {
	rstms.OptionSwitch(cobraCmd, name, flag, description)
}

func OptionString(cobraCmd CobraCommand, name, flag, defaultValue, description string) {
	rstms.OptionString(cobraCmd, name, flag, defaultValue, description)
}

func OptionStringSlice(cobraCmd CobraCommand, name, flag string, defaultValue []string, description string) {
	rstms.OptionStringSlice(cobraCmd, name, flag, defaultValue, description)
}

func OptionInt(cobraCmd CobraCommand, name, flag string, defaultValue int, description string) {
	rstms.OptionInt(cobraCmd, name, flag, defaultValue, description)
}

func CobraAddCommand(cobraRootCmd, parentCmd, cobraCmd CobraCommand) {
	rstms.CobraAddCommand(cobraRootCmd, parentCmd, cobraCmd)
}

func CobraInit(cobraRootCmd CobraCommand) {
	rstms.CobraInit(cobraRootCmd)
}

func Init(name, version, configFile string) {
	rstms.Init(name, version, configFile)
}

func Shutdown() {
	rstms.Shutdown()
}

func ProgramName() string {
	return rstms.ProgramName()
}

func ProgramVersion() string {
	return rstms.ProgramVersion()
}

func ConfigDir() string {
	return rstms.ConfigDir()
}

func CheckErr(err error) {
	rstms.CheckErr(err)
}

func FormatJSON(v any) string {
	return rstms.FormatJSON(v)
}

func ConfigString(header bool) string {
	return rstms.ConfigString(header)
}

func FormatYAML(value any) string {
	return rstms.FormatYAML(value)
}

func ConfigInit(allowClobber bool) string {
	return rstms.ConfigInit(allowClobber)
}

func ConfigEdit() {
	rstms.ConfigEdit()
}

func AppendConfig(filename string) error {
	return rstms.AppendConfig(filename)
}

func Confirm(prompt string) bool {
	return rstms.Confirm(prompt)
}

func Fatal(err error) error {
	return rstms.Fatal(err)
}

func Fatalf(format string, args ...interface{}) error {
	return rstms.Fatalf(format, args...)
}

func Warning(format string, args ...interface{}) {
	rstms.Warning(format, args...)
}

func HexDump(data []byte) string {
	return rstms.HexDump(data)
}

func GetHostnameDetail() (string, string, string, error) {
	return rstms.GetHostnameDetail()
}

func HostShortname() (string, error) {
	return rstms.HostShortname()
}

func HostDomain() (string, error) {
	return rstms.HostDomain()
}

func HostFQDN() (string, error) {
	return rstms.HostFQDN()
}

func IsDir(path string) bool {
	return rstms.IsDir(path)
}

func IsFile(pathname string) bool {
	return rstms.IsFile(pathname)
}

func TildePath(path string) (string, error) {
	return rstms.TildePath(path)
}

func NewSendmail(hostname string, port int, username, password, CAFile string) (Sendmail, error) {
	return rstms.NewSendmail(hostname, port, username, password, CAFile)
}

func Expand(value string) string {
	return rstms.Expand(value)
}

func ViperKey(key string) string {
	return rstms.ViperKey(key)
}

func ViperGet(key string) any {
	return rstms.ViperGet(key)
}

func ViperGetBool(key string) bool {
	return rstms.ViperGetBool(key)
}

func ViperGetString(key string) string {
	return rstms.ViperGetString(key)
}

func ViperGetStringSlice(key string) []string {
	return rstms.ViperGetStringSlice(key)
}

func ViperGetStringMapString(key string) map[string]string {
	return rstms.ViperGetStringMapString(key)
}

func ViperGetStringMap(key string) map[string]any {
	return rstms.ViperGetStringMap(key)
}

func ViperGetInt(key string) int {
	return rstms.ViperGetInt(key)
}

func ViperGetInt64(key string) int64 {
	return rstms.ViperGetInt64(key)
}

func ViperSet(key string, value any) {
	rstms.ViperSet(key, value)
}

func ViperSetDefault(key string, value any) {
	rstms.ViperSetDefault(key, value)
}
