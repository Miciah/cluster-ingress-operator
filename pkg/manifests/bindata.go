// Code generated by go-bindata. DO NOT EDIT.
// sources:
// assets/router/cluster-role-binding.yaml (329B)
// assets/router/cluster-role.yaml (788B)
// assets/router/deployment.yaml (1.69kB)
// assets/router/metrics/cluster-role-binding.yaml (285B)
// assets/router/metrics/cluster-role.yaml (259B)
// assets/router/metrics/role-binding.yaml (297B)
// assets/router/metrics/role.yaml (291B)
// assets/router/service-account.yaml (213B)
// assets/router/service-cloud.yaml (628B)
// assets/router/service-internal.yaml (427B)

package manifests

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsRouterClusterRoleBindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x84\x8f\x31\x4e\xc4\x40\x0c\x45\xfb\x39\x85\x25\xea\x0c\xa2\x43\xd3\x01\x37\x58\x24\x7a\xef\xc4\xbb\x31\x49\xec\xc8\xf6\xa4\xe0\xf4\x28\x4a\x44\xc3\x4a\x29\x2d\xf9\xbf\xff\xfe\x13\xbc\xb3\xf4\x0e\x31\x10\x98\xb6\x20\x03\xd3\x89\x20\x14\x38\x1c\x3e\xc9\x56\xae\x04\x6f\xb5\x6a\x93\xc8\x69\x64\xe9\x0b\x7c\x4c\xcd\x83\xec\xa2\x13\x6d\x71\x96\x7b\xc2\x85\xbf\xc8\x9c\x55\x0a\xd8\x15\x6b\xc6\x16\x83\x1a\xff\x60\xb0\x4a\x1e\x5f\x3d\xb3\x3e\xaf\x2f\x69\xa6\xc0\x1e\x03\x4b\x02\x10\x9c\xa9\x80\x2e\x24\x3e\xf0\x2d\x3a\x96\xbb\x91\x7b\xb7\x9b\x24\x6f\xd7\x6f\xaa\xe1\x25\x75\xb0\x17\x1f\x3e\x87\xce\x1f\xe1\xf8\xdf\x4f\x5f\xb0\x3e\xa2\xa6\x6d\xd8\x85\x6e\x5b\xf1\xbf\x19\xe7\x32\x27\xf0\xdf\x00\x00\x00\xff\xff\x83\x13\xa9\xa6\x49\x01\x00\x00")

func assetsRouterClusterRoleBindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterClusterRoleBindingYaml,
		"assets/router/cluster-role-binding.yaml",
	)
}

func assetsRouterClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRouterClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/cluster-role-binding.yaml", size: 329, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x12, 0x9a, 0xeb, 0xd3, 0x79, 0x1a, 0xa6, 0x75, 0xb5, 0xda, 0x10, 0x70, 0xf9, 0x80, 0xd1, 0x51, 0x55, 0x98, 0x4d, 0x11, 0x98, 0x79, 0x5c, 0x46, 0x99, 0xa3, 0x39, 0x68, 0xe9, 0xa2, 0x72, 0x56}}
	return a, nil
}

var _assetsRouterClusterRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x92\x31\x6f\xe3\x30\x0c\x85\x77\xfd\x0a\x22\x37\xdb\xc1\x6d\x07\xaf\x37\xdc\x76\x43\x51\x74\xa7\x65\xa6\x66\xed\x88\x02\x49\x39\x6d\x7f\x7d\x61\x3b\x29\x82\x24\x45\x9b\xcd\xcf\x22\xbf\x47\x3e\xe9\x17\xfc\x1d\x8b\x39\x29\x58\x94\x4c\x1d\xa8\x8c\x04\x3b\x51\x50\x29\x4e\x6a\x35\x3c\xf6\x6c\x60\xbd\x94\xb1\x83\x96\x00\x0d\x94\xcc\x95\xa3\xf3\xb4\xc8\x2c\x66\xdc\x8e\x54\x87\x81\x53\xd7\x9c\x88\x0f\x32\x52\xc0\xcc\x4f\xa4\xc6\x92\x1a\xd0\x16\x63\x8d\xc5\x7b\x51\x7e\x47\x67\x49\xf5\xf0\xc7\x6a\x96\xed\xf4\x3b\xec\xc9\xb1\x43\xc7\x26\x00\x24\xdc\x53\x03\x92\x29\x59\xcf\x3b\xaf\x38\x3d\x2b\x99\x55\xeb\x48\x41\xcb\x48\xd6\x84\x0a\x30\xf3\x3f\x95\x92\x6d\x6e\xaa\x60\xb3\x09\x30\xcf\x26\x45\x23\x1d\xff\x51\xea\xb2\x70\x72\x5b\xd4\x0c\xb6\x8c\x91\x56\x69\xa4\x13\xaf\x62\x22\x6d\x8f\x2d\x23\x9b\x2f\x1f\x07\xf4\xd8\x87\x6b\x9f\x79\x05\x4a\xce\xf1\x7c\x87\x6b\x6b\x97\x81\x92\xd2\xc4\x74\xb8\x70\x88\x4a\xe8\xf4\x05\xf9\x32\x9c\x6b\xb0\x95\xf6\x85\xa2\x63\x8c\x64\x76\x9f\xc1\x92\x60\xfd\x99\xec\x4d\xfc\x52\x73\x6f\x26\x3f\x07\x6f\xcd\xd1\xcb\x05\xbf\xe4\xee\xf6\xc0\x46\xb1\x28\xfb\xdb\x37\xe8\x53\x59\x94\xe4\xf4\xea\x51\x92\xb9\xe2\xf1\xde\xcf\x7d\x8c\xce\x9a\xff\xcf\xcf\x61\x3d\xe8\xc5\x3c\x91\x1f\x44\x87\xf0\x11\x00\x00\xff\xff\xad\x45\xb2\xc3\x14\x03\x00\x00")

func assetsRouterClusterRoleYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterClusterRoleYaml,
		"assets/router/cluster-role.yaml",
	)
}

func assetsRouterClusterRoleYaml() (*asset, error) {
	bytes, err := assetsRouterClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/cluster-role.yaml", size: 788, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xa, 0xc4, 0x4e, 0xa2, 0x9d, 0x1b, 0xd7, 0x35, 0xbf, 0x95, 0x95, 0xb4, 0x17, 0x87, 0x56, 0xb, 0x12, 0xf, 0xb6, 0x3e, 0x51, 0xe5, 0x5a, 0xd0, 0x66, 0x3a, 0x4d, 0x36, 0xc9, 0x93, 0x90, 0xb2}}
	return a, nil
}

var _assetsRouterDeploymentYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x94\x4b\x6f\x22\x39\x10\xc7\xef\x7c\x8a\x52\x38\xb3\xe4\xb5\xd1\xae\x6f\x08\x3a\x2b\xa4\x25\x69\x41\x27\x57\xe4\x71\x17\x60\xc5\xaf\xa9\xaa\x26\x62\x3e\xfd\xc8\x3c\x26\x34\x21\x51\xe6\x36\x7d\x6a\xb9\xfe\xfe\xd5\xdf\xe5\x2a\x77\x61\x84\xc9\xc5\x8d\xc7\x20\xf0\x6a\x65\x05\x35\x2e\x74\xe3\x04\xd6\xda\x35\xc8\x9d\x2e\x0c\x5d\xc3\x82\x04\xe3\xb0\x24\x64\x06\x4e\x68\xec\xc2\x9a\xbd\x02\x34\x21\xe8\x94\x9c\xc5\x1a\xb4\x00\x35\x41\xac\xc7\xbf\x3a\x2f\x36\xd4\xea\x08\xdf\xd1\xc9\x3e\x23\xb1\x8d\x41\xe5\x0d\xdc\x5f\x5f\x75\x3c\x8a\xae\xb5\x68\xd5\x01\xe8\xc2\x83\xf6\x08\x96\x81\x51\x5a\x28\x80\xa0\x3d\x72\xd2\x06\x15\xc4\x84\x81\x57\x76\x21\x3d\xbb\x73\xd4\xc9\x8e\x32\x40\xd0\x27\xa7\x05\xf3\x3f\xc0\x61\x75\xfb\x8f\xb4\xb6\x06\x07\xc6\xc4\x26\x48\x4e\xa3\x80\x62\x23\x48\x7b\x41\x17\x42\xac\x71\x86\x0e\x8d\x44\x3a\xef\x21\x7f\x89\x6c\x24\x2b\x9b\xa1\xd3\xcc\x3b\x0e\x6f\x58\xd0\xf7\xcc\xae\x4c\x3d\x43\x56\xac\xd1\x6e\xbf\xc1\xc4\x20\xda\x06\x24\x3e\x78\x01\xe8\x6d\x8f\x73\xe2\x60\xe7\xc2\x7a\xbd\xfc\xa0\x04\x87\x6f\x2b\x29\x1b\xe7\xca\xe8\xac\xd9\x28\x18\x2f\x1e\xa2\x94\x84\x9c\x8b\xfc\xa6\x4b\x91\xe4\x28\xe9\x5b\xda\x95\x48\x3a\x5a\x3e\xf2\x58\x46\x12\x05\xff\x5c\xb6\xa2\x89\xa2\x44\x13\x9d\x82\x6a\x58\x7e\x80\xe3\xcf\x78\xb7\xb7\x37\xbf\x05\xf4\x28\x64\xcd\xa7\xc8\xab\x7f\x6f\xee\xbe\xc4\xec\xc2\x04\x69\x79\xd2\x99\x6f\x61\x0c\x6b\xd5\x52\xb3\x68\x61\x68\x18\x29\x7b\x01\x1d\x6a\x48\x9a\xf9\x35\x52\xbd\x6d\xf4\x25\x06\x24\x2d\x2d\xe0\x99\x23\xcc\xaa\x41\x35\x9b\x97\x8f\xd3\xaa\xe5\x72\x3b\x31\x0a\x2e\xb2\xfd\x8b\x33\xdb\xa6\x8f\x4f\x55\x31\x9d\xcf\x8a\xe9\xf3\x78\x58\xcc\x1f\x06\x93\x62\x56\x0e\x86\xc5\x39\xc8\xfb\x31\x78\xcf\x1b\x15\xf7\x83\xa7\xff\xab\xf9\xb0\x98\x56\xe3\xfb\xf1\x70\x50\x15\xf3\xd1\x78\x7a\x0e\xd7\x47\x31\xfd\xf4\x62\xfb\xe2\xb8\x9f\xc8\xae\xb5\x1c\x1f\xcc\xd9\x35\x06\x64\x2e\x29\x7e\x43\xd5\x02\xd8\x60\xc5\x6a\x37\x42\xa7\x37\x33\x34\x31\xd4\xac\xe0\xaa\xdd\x43\xb9\x47\xfe\x43\x69\x6f\x04\x48\x5a\x56\x0a\xfa\x2b\xd4\x4e\x56\x3f\x4e\x83\xe7\x6e\x9a\x50\xd7\xf6\xcf\x30\xc2\xb1\x21\x83\xdc\x46\x11\x7e\x6f\x90\x85\x4f\x13\x98\xd4\x64\x2f\x97\xfe\x64\xdd\xa3\x8f\xb4\x51\x70\xfd\xf7\xdd\xc4\x1e\xc5\xd6\xd1\x35\x1e\x27\xf9\xb5\x3a\x99\x61\x9f\xd7\xca\x9d\xdf\xcf\xef\x0c\xf6\x5d\xb0\x7f\xd1\x7b\x06\x49\xf2\xc3\x7d\xaa\xca\x35\x7d\x0c\x6e\xa3\x40\xa8\x39\x84\x76\x06\x7e\xe5\xee\x7d\x81\xc5\x68\xa8\x5d\xda\xbd\x7a\x12\x6b\x54\x70\x7b\x7d\xd9\x1a\xb5\xd9\x56\xfe\xfe\xc9\xef\xed\x86\xf4\x67\x00\x00\x00\xff\xff\x16\xb5\x79\x00\x9a\x06\x00\x00")

func assetsRouterDeploymentYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterDeploymentYaml,
		"assets/router/deployment.yaml",
	)
}

func assetsRouterDeploymentYaml() (*asset, error) {
	bytes, err := assetsRouterDeploymentYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/deployment.yaml", size: 1690, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xc9, 0x9a, 0x75, 0x63, 0xed, 0x88, 0x74, 0x2b, 0x2, 0xe6, 0xd7, 0xac, 0xe0, 0x86, 0xb8, 0xea, 0xb0, 0xaf, 0x1, 0xb8, 0xe8, 0x2, 0xa4, 0x2b, 0x45, 0xc, 0xc4, 0x28, 0xc0, 0x39, 0x5a, 0x19}}
	return a, nil
}

var _assetsRouterMetricsClusterRoleBindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x8f\xc1\x4a\xc4\x40\x0c\x86\xef\xf3\x14\x79\x81\x56\xbc\x2d\x73\x53\x0f\xde\x57\xf0\x9e\x9d\xa6\x36\xb6\x93\x0c\x49\xa6\x07\x9f\x5e\x8a\x22\xc2\x42\xaf\x81\x7c\xdf\xff\xad\x2c\x53\x86\x97\xad\x7b\x90\x5d\x75\xa3\x67\x96\x89\xe5\x23\x61\xe3\x77\x32\x67\x95\x0c\x76\xc3\x32\x62\x8f\x45\x8d\xbf\x30\x58\x65\x5c\x2f\x3e\xb2\x3e\xec\x8f\xa9\x52\xe0\x84\x81\x39\x01\x08\x56\xca\x60\xda\x83\x6c\xa8\x2a\x1c\x6a\x07\xcc\xfb\xed\x93\x4a\x78\x4e\x03\xfc\x18\xdf\xc8\x76\x2e\xf4\x54\x8a\x76\x89\xbf\xd7\x66\x5a\x29\x16\xea\x3e\xac\x17\xff\x3d\x7b\xc3\x42\x19\xb4\x91\xf8\xc2\x73\xfc\x27\x9b\x6e\x74\xa5\xf9\x90\xdf\xa5\x9c\x0c\x02\xc0\xc6\xaf\xa6\xbd\x9d\xd4\xa5\xef\x00\x00\x00\xff\xff\x7f\xc0\x4a\x40\x1d\x01\x00\x00")

func assetsRouterMetricsClusterRoleBindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterMetricsClusterRoleBindingYaml,
		"assets/router/metrics/cluster-role-binding.yaml",
	)
}

func assetsRouterMetricsClusterRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRouterMetricsClusterRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/metrics/cluster-role-binding.yaml", size: 285, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xa2, 0xce, 0x4e, 0xd1, 0x37, 0xde, 0x79, 0x91, 0x5c, 0x71, 0xd1, 0x88, 0x1b, 0xdb, 0xaf, 0x1, 0xe5, 0x8c, 0x81, 0xb3, 0xfd, 0x30, 0xe3, 0x5d, 0xb0, 0x59, 0x8b, 0x2a, 0x47, 0xf9, 0xa0, 0xdf}}
	return a, nil
}

var _assetsRouterMetricsClusterRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\xce\xb1\x4e\xc4\x40\x0c\x84\xe1\x7e\x9f\xc2\x12\x75\x72\xa2\x43\x69\x29\xe8\x29\xe8\x9d\xec\x70\xb1\x2e\x6b\xaf\x6c\xef\x49\xf0\xf4\xe8\x44\x90\xa8\xe7\x93\xfe\x79\xa2\xd7\x63\x44\xc2\xc9\xed\x40\x90\x02\x15\x95\xd6\x2f\xea\x6e\x0d\xb9\x63\x04\xa5\x51\x6c\xce\x1d\xe4\x36\x1e\xb6\x21\x5d\xb6\x20\x68\xed\x26\x9a\x85\xbb\x7c\xc0\x43\x4c\x17\xf2\x95\xb7\x99\x47\xee\xe6\xf2\xcd\x29\xa6\xf3\xed\x25\x66\xb1\xcb\xfd\xb9\xdc\x44\xeb\xf2\xd7\x7c\xb7\x03\xa5\x21\xb9\x72\xf2\x52\x88\x94\x1b\x96\x33\x32\x35\x53\x49\x73\xd1\x6b\xf1\x71\x20\x96\x32\x11\x77\x79\x73\x1b\x3d\x1e\x7a\xfa\x95\xb3\x75\x68\xec\xf2\x99\xb3\x58\x21\x72\x84\x0d\xdf\xf0\xdf\x78\x5c\xce\xcf\x85\xe8\x0e\x5f\xcf\xf1\x8a\x2c\x3f\x01\x00\x00\xff\xff\x4f\xd5\xdf\xe0\x03\x01\x00\x00")

func assetsRouterMetricsClusterRoleYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterMetricsClusterRoleYaml,
		"assets/router/metrics/cluster-role.yaml",
	)
}

func assetsRouterMetricsClusterRoleYaml() (*asset, error) {
	bytes, err := assetsRouterMetricsClusterRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/metrics/cluster-role.yaml", size: 259, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x98, 0x77, 0x73, 0x9c, 0x6, 0x33, 0xf2, 0x91, 0x6f, 0x3b, 0x35, 0x49, 0xf3, 0xa5, 0xfc, 0x1d, 0x2e, 0x2e, 0xa6, 0x5b, 0x95, 0xa6, 0x7e, 0x8d, 0xfe, 0x7e, 0xf4, 0x62, 0x30, 0xa5, 0x37, 0x61}}
	return a, nil
}

var _assetsRouterMetricsRoleBindingYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\xce\x31\x4e\xc5\x40\x0c\x04\xd0\x7e\x4f\xe1\x0b\x24\x88\xee\x6b\x3b\x68\xe8\x3f\x12\xbd\xb3\x71\x12\x93\xac\xbd\xb2\xbd\x29\x38\x3d\x42\x8a\x44\x05\xd2\x6f\x47\x33\x9a\x87\x8d\x3f\xc8\x9c\x55\x32\xd8\x84\x65\xc4\x1e\x9b\x1a\x7f\x61\xb0\xca\xb8\xdf\x7c\x64\x7d\x3a\x9f\xd3\xce\x32\x67\xb8\xeb\x41\xaf\x2c\x33\xcb\x9a\x2a\x05\xce\x18\x98\x13\x80\x60\xa5\x0c\xcd\xb4\x52\x6c\xd4\x7d\xd8\x6f\x7e\xc5\xde\xb0\x50\x06\x6d\x24\xbe\xf1\x12\x03\xcb\x6a\xe4\x9e\x4c\x0f\xba\xd3\xf2\x33\xc7\xc6\x6f\xa6\xbd\xfd\x63\x48\x00\xbf\x84\xbf\x1e\xbd\x4f\x9f\x54\xc2\x73\x1a\xae\xf6\x3b\xd9\xc9\x85\x5e\x4a\xd1\x2e\xf1\xa0\xb4\xaa\x70\xa8\xb1\xac\x90\xbe\x03\x00\x00\xff\xff\x15\x9f\x30\x56\x29\x01\x00\x00")

func assetsRouterMetricsRoleBindingYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterMetricsRoleBindingYaml,
		"assets/router/metrics/role-binding.yaml",
	)
}

func assetsRouterMetricsRoleBindingYaml() (*asset, error) {
	bytes, err := assetsRouterMetricsRoleBindingYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/metrics/role-binding.yaml", size: 297, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xff, 0xef, 0x8, 0xfb, 0x1f, 0xa3, 0xc7, 0xfb, 0xbc, 0x6, 0x78, 0xad, 0x0, 0x28, 0x90, 0xc8, 0xe8, 0xf5, 0x7d, 0xf8, 0xd0, 0xeb, 0x52, 0xf, 0xd4, 0x81, 0xce, 0x69, 0xb8, 0x8c, 0x26, 0x8c}}
	return a, nil
}

var _assetsRouterMetricsRoleYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8e\xb1\x6e\xeb\x30\x0c\x45\x77\x7d\x05\x91\x37\x3b\x0f\xdd\x02\xfd\x40\xf7\x0e\xdd\x19\xe9\x36\x26\x62\x8b\x02\x49\xb9\x68\xbf\xbe\x88\xe3\x02\x9d\x78\xef\x01\xc1\xc3\x7f\xf4\xa6\x0b\x9c\x1a\x50\x51\xe9\xfa\x45\xdd\x74\x45\xcc\x18\x4e\xa1\xe4\xc5\xb8\x83\x4c\x47\xc0\x68\x45\x98\x14\x27\xb4\xda\x55\x5a\x24\xee\xf2\x0e\x73\xd1\x96\xc9\xae\x5c\xce\x3c\x62\x56\x93\x6f\x0e\xd1\x76\xbe\x5f\xfc\x2c\xfa\x7f\x7b\x49\x77\x69\x35\xef\xae\xb4\x22\xb8\x72\x70\x4e\x44\x8d\x57\xe4\x3f\xca\xe9\x7e\xf1\x03\x7b\xe7\x82\x4c\xda\xd1\x7c\x96\x8f\x98\xa4\xdd\x0c\xee\xc9\xc6\x02\xcf\x69\x22\xee\xf2\x6a\x3a\xba\x3f\x2e\x4d\x74\x3a\x25\x22\x83\xeb\xb0\x82\x83\x39\x6c\x93\x02\xdf\xcb\xef\xd7\xcf\xd6\xb5\x3e\xc2\x06\xbb\x1e\xcb\x37\xc4\x3e\x17\xf1\x67\xf8\xe4\x28\x73\xfa\x09\x00\x00\xff\xff\x67\x78\x6f\x08\x23\x01\x00\x00")

func assetsRouterMetricsRoleYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterMetricsRoleYaml,
		"assets/router/metrics/role.yaml",
	)
}

func assetsRouterMetricsRoleYaml() (*asset, error) {
	bytes, err := assetsRouterMetricsRoleYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/metrics/role.yaml", size: 291, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x57, 0xe3, 0x68, 0x2c, 0xfd, 0xf1, 0x27, 0x9d, 0xa0, 0x3b, 0x10, 0x3e, 0xca, 0x3b, 0x76, 0x39, 0xf4, 0xb1, 0x37, 0x7b, 0xa3, 0xa7, 0x11, 0xc0, 0x6, 0x4b, 0x47, 0xbb, 0x93, 0x4b, 0xb7, 0xc2}}
	return a, nil
}

var _assetsRouterServiceAccountYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x2c\xce\xb1\x4e\xc4\x30\x10\x84\xe1\xde\x4f\x31\xd2\xd5\x9c\x44\xeb\x8e\x92\x16\x24\x7a\xb3\x99\xbb\x5b\x91\x78\xcd\xee\x3a\x88\xb7\x47\x41\x29\xa7\x98\x5f\xdf\x05\x2f\x22\x36\x7b\xe2\x66\x0e\xb7\x99\xf4\x80\x38\x5b\x72\xc1\xe7\x2f\xf2\x41\xd8\xa0\xb7\x34\xbf\xe2\x35\xf1\xa3\xeb\x0a\xe7\xf7\x54\x27\x64\x9d\x91\x74\x84\xd8\xe0\x52\x2e\x18\xf4\x4d\x23\xd4\x7a\xc0\xb9\xfe\x57\xd2\xf0\x76\x84\x31\xdc\x84\x11\xda\xef\xd7\xf2\xa5\x7d\xa9\x78\xa7\xef\x2a\x3c\x0d\xa5\x0d\xfd\xa0\x1f\xef\x8a\xfd\xb9\x6c\xcc\xb6\xb4\x6c\xb5\x00\xbd\x6d\xac\x27\xf0\x9c\x31\x9a\xb0\x1e\xba\x1e\x0f\xbd\xe5\x93\xf6\xbb\x33\xa2\xfc\x05\x00\x00\xff\xff\x33\xdc\xda\x8c\xd5\x00\x00\x00")

func assetsRouterServiceAccountYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterServiceAccountYaml,
		"assets/router/service-account.yaml",
	)
}

func assetsRouterServiceAccountYaml() (*asset, error) {
	bytes, err := assetsRouterServiceAccountYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/service-account.yaml", size: 213, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xd0, 0xe3, 0x6, 0x3a, 0x88, 0x2e, 0x33, 0xe3, 0x24, 0xf0, 0xf0, 0xe9, 0x43, 0xc8, 0x46, 0x6c, 0x60, 0x9, 0x69, 0x84, 0x3, 0xd8, 0xc3, 0x80, 0xb, 0xab, 0x37, 0x13, 0xce, 0xf2, 0xeb, 0x60}}
	return a, nil
}

var _assetsRouterServiceCloudYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x7c\x90\x41\x6b\x14\x41\x10\x85\xef\xfd\x2b\x1e\xe4\x9c\xa0\x98\x83\xcc\x31\x39\x09\x41\x16\x5c\xbc\x57\x7a\x6a\x76\x9a\xf4\x54\x35\x55\x35\xab\xfb\xef\xa5\x7b\x37\xa0\x28\x1e\xe7\x31\xfd\xd5\x7b\xdf\x1d\x5e\x94\x66\x3c\x51\x25\xc9\x6c\xf8\xc6\x76\x2e\x99\x11\x8a\x56\x29\x33\x8a\x60\x31\x95\x80\x2e\x88\x95\x61\xba\x07\x5b\x8f\x73\xd5\x7d\x06\xcb\xb9\x98\xca\xc6\x12\xfe\x90\xee\xf0\x5c\x77\xef\x3f\x7c\x91\x93\xb1\x3b\xbc\x71\x2e\x4b\xc9\x38\x53\xdd\xd9\x41\xc6\xa0\xd6\x6a\xe1\x19\x14\xb0\x5d\xa2\x6c\xfc\x90\xde\x8a\xcc\xd3\xfb\xf9\x44\xad\x7c\x67\xf3\xa2\x32\xe1\xfc\x31\x6d\x1c\x34\x53\xd0\x94\x80\x3b\x7c\xa5\x8d\x51\x1c\xce\xf1\x07\x02\x10\xda\xd8\x1b\x65\x9e\xa0\x8d\xc5\xd7\xb2\xc4\x7d\xb9\x36\x49\x40\xa5\x57\xae\xde\x21\xe8\x1d\xa6\xdb\x98\xd4\x3b\xf6\x34\x2e\x8d\xa7\x21\xe4\xdd\x47\x02\x9c\x2b\xe7\x50\xfb\xfb\x59\xef\x72\x5c\x8b\x83\xaa\x2b\x56\xf2\x21\x88\x97\x85\xf3\xd0\xb5\x91\xbd\x15\x39\xe1\xe5\x09\x4d\xb5\x22\xc8\x4e\x1c\x0e\x72\xec\xb2\x32\xd5\x58\x2f\xf8\xb1\xb2\x40\x74\xc0\x6e\x6e\x9b\xce\x57\x4f\xcd\xd8\xb9\xab\x17\x10\x44\x67\xc6\x2b\xaf\x45\xe6\x71\xc7\xaf\xaa\xfa\x6c\xfe\x19\x6c\x42\xf5\x68\xb4\x2c\x25\x1f\xb4\x96\x7c\xe9\x43\x32\xd5\x04\x34\xb5\x18\xab\xef\x87\xa0\x09\x6b\x44\x1b\x6b\x9a\x69\x68\xd6\x3a\xe1\xf8\x7c\xb8\x26\x6a\x31\xe1\xf3\x87\xf1\x71\x2d\x7c\x18\xd1\xed\xcd\xef\x08\xff\x2f\xe3\xf1\xf1\xd3\x3f\x21\x9e\x7e\x05\x00\x00\xff\xff\x14\xac\xd6\xf5\x74\x02\x00\x00")

func assetsRouterServiceCloudYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterServiceCloudYaml,
		"assets/router/service-cloud.yaml",
	)
}

func assetsRouterServiceCloudYaml() (*asset, error) {
	bytes, err := assetsRouterServiceCloudYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/service-cloud.yaml", size: 628, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x8c, 0x35, 0x3d, 0xd7, 0x8, 0xf9, 0xba, 0x0, 0x54, 0xbd, 0x2a, 0xeb, 0x98, 0x83, 0x6f, 0x28, 0x5e, 0xda, 0xd8, 0xa9, 0x45, 0x65, 0xe2, 0x35, 0x98, 0xd0, 0x6, 0x64, 0xc4, 0x82, 0x36, 0x14}}
	return a, nil
}

var _assetsRouterServiceInternalYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\xce\x4f\x4b\xc3\x40\x10\x05\xf0\xfb\x7e\x8a\x07\x3d\x2b\x96\x16\xd1\xbd\x49\x4f\xbd\x48\x40\xf1\x3e\x24\xd3\x76\x30\xfb\x87\x99\x49\xc4\x6f\x2f\x49\x2d\xc6\x9e\x7a\xdc\x9d\x37\xbf\x79\x2b\xec\xfa\xc1\x9c\x15\x6f\xac\xa3\xb4\x8c\x2f\xf1\x13\x3a\x3e\xd0\xd0\x3b\x46\xea\x07\xb6\xf0\x97\xda\xe7\xa3\xb2\x19\xac\x72\x2b\x07\x69\x41\x39\x17\x27\x97\x92\x0d\xa4\x0c\xaa\xb5\x17\xee\x40\x0e\x1d\xb2\x4b\xe2\xfb\xf0\x29\xb9\x8b\x97\x03\x81\xaa\x7c\xb0\x9a\x94\x1c\x31\xae\x43\x62\xa7\x8e\x9c\x62\x00\x56\x78\xa5\xc4\xa0\xdc\xe1\xe5\xca\x35\xf6\x7f\xe6\x54\x60\x5a\xf1\xef\xca\xf1\x52\x6f\xdf\x04\xa0\x16\x75\x9b\x46\x77\xc8\x94\x38\xe2\xe4\x5e\x03\x70\x9e\x44\x3c\x3d\x9c\x1f\x5a\xbc\xb4\xa5\x8f\x78\xdf\x35\xf3\x8f\x93\x1e\xd9\x9b\x39\xf4\xbb\xb3\x24\x6c\x61\x6c\xb7\x9b\x1b\x11\x5b\x28\x89\x5d\xa5\x5d\x3a\xeb\xe7\xcd\xe3\x0d\xd0\x1c\xfb\x09\x00\x00\xff\xff\x21\xc4\xfd\xf7\xab\x01\x00\x00")

func assetsRouterServiceInternalYamlBytes() ([]byte, error) {
	return bindataRead(
		_assetsRouterServiceInternalYaml,
		"assets/router/service-internal.yaml",
	)
}

func assetsRouterServiceInternalYaml() (*asset, error) {
	bytes, err := assetsRouterServiceInternalYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/router/service-internal.yaml", size: 427, mode: os.FileMode(420), modTime: time.Unix(1, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xc5, 0xc5, 0x46, 0xc8, 0x1c, 0x5f, 0x62, 0x59, 0x73, 0x14, 0x59, 0xc1, 0x9b, 0xf5, 0x7a, 0x5, 0x18, 0xf7, 0x14, 0x2, 0x0, 0x2d, 0x2, 0x23, 0x8a, 0x4d, 0x49, 0x3b, 0xf3, 0xbd, 0xfe, 0x17}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/router/cluster-role-binding.yaml": assetsRouterClusterRoleBindingYaml,

	"assets/router/cluster-role.yaml": assetsRouterClusterRoleYaml,

	"assets/router/deployment.yaml": assetsRouterDeploymentYaml,

	"assets/router/metrics/cluster-role-binding.yaml": assetsRouterMetricsClusterRoleBindingYaml,

	"assets/router/metrics/cluster-role.yaml": assetsRouterMetricsClusterRoleYaml,

	"assets/router/metrics/role-binding.yaml": assetsRouterMetricsRoleBindingYaml,

	"assets/router/metrics/role.yaml": assetsRouterMetricsRoleYaml,

	"assets/router/service-account.yaml": assetsRouterServiceAccountYaml,

	"assets/router/service-cloud.yaml": assetsRouterServiceCloudYaml,

	"assets/router/service-internal.yaml": assetsRouterServiceInternalYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": {nil, map[string]*bintree{
		"router": {nil, map[string]*bintree{
			"cluster-role-binding.yaml": {assetsRouterClusterRoleBindingYaml, map[string]*bintree{}},
			"cluster-role.yaml":         {assetsRouterClusterRoleYaml, map[string]*bintree{}},
			"deployment.yaml":           {assetsRouterDeploymentYaml, map[string]*bintree{}},
			"metrics": {nil, map[string]*bintree{
				"cluster-role-binding.yaml": {assetsRouterMetricsClusterRoleBindingYaml, map[string]*bintree{}},
				"cluster-role.yaml":         {assetsRouterMetricsClusterRoleYaml, map[string]*bintree{}},
				"role-binding.yaml":         {assetsRouterMetricsRoleBindingYaml, map[string]*bintree{}},
				"role.yaml":                 {assetsRouterMetricsRoleYaml, map[string]*bintree{}},
			}},
			"service-account.yaml":  {assetsRouterServiceAccountYaml, map[string]*bintree{}},
			"service-cloud.yaml":    {assetsRouterServiceCloudYaml, map[string]*bintree{}},
			"service-internal.yaml": {assetsRouterServiceInternalYaml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
