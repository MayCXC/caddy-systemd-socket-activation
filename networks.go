package networks

import (
	"context"
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

func init() {
	caddy.RegisterNetwork("sd", getListener)
	caddy.RegisterNetwork("sdgram", getListener)
	caddyhttp.RegisterNetworkHTTP3("sdgram", "sdgram")
}

var (
	nameToFiles    map[string][]int
	nameToFilesErr error
	nameToFilesMu  sync.Mutex
)

func sdListenFds() (map[string][]int, error) {
	nameToFilesMu.Lock()
	defer nameToFilesMu.Unlock()

	if nameToFilesErr != nil {
		return nil, nameToFilesErr
	}

	if nameToFiles != nil {
		return nameToFiles, nil
	}

	const lnFdsStart = 3

	lnPid, ok := os.LookupEnv("LISTEN_PID")
	if !ok {
		nameToFilesErr = errors.New("LISTEN_PID is unset.")
		return nil, nameToFilesErr
	}

	pid, err := strconv.ParseUint(lnPid, 0, strconv.IntSize)
	if err != nil {
		nameToFilesErr = err
		return nil, nameToFilesErr
	}

	if pid != uint64(os.Getpid()) {
		nameToFilesErr = fmt.Errorf("LISTEN_PID does not match pid: %d != %d", pid, os.Getpid())
		return nil, nameToFilesErr
	}

	lnFds, ok := os.LookupEnv("LISTEN_FDS")
	if !ok {
		nameToFilesErr = errors.New("LISTEN_FDS is unset.")
		return nil, nameToFilesErr
	}

	fds, err := strconv.ParseUint(lnFds, 0, strconv.IntSize)
	if err != nil {
		nameToFilesErr = err
		return nil, nameToFilesErr
	}

	lnFdnames, ok := os.LookupEnv("LISTEN_FDNAMES")
	if !ok {
		nameToFilesErr = errors.New("LISTEN_FDNAMES is unset.")
		return nil, nameToFilesErr
	}

	fdNames := strings.Split(lnFdnames, ":")
	if fds != uint64(len(fdNames)) {
		nameToFilesErr = fmt.Errorf("LISTEN_FDS does not match LISTEN_FDNAMES length: %d != %d", fds, len(fdNames))
		return nil, nameToFilesErr
	}

	nameToFiles = make(map[string][]int, len(fdNames))
	for index, name := range fdNames {
		nameToFiles[name] = append(nameToFiles[name], lnFdsStart+index)
	}

	return nameToFiles, nil
}

func getListener(ctx context.Context, network, addr string, cfg net.ListenConfig) (any, error) {
	lnFds, err := sdListenFds()
	if err != nil {
		return nil, err
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	name, index, li := host, "0", strings.LastIndex(host, "/")
	if li >= 0 {
		name = host[:li]
		index = host[li+1:]
	}

	i, err := strconv.ParseUint(index, 0, strconv.IntSize)
	if err != nil {
		return nil, err
	}

	files, ok := lnFds[name]
	if !ok {
		return nil, fmt.Errorf("invalid listen fd name: %s", name)
	}

	if uint64(len(files)) <= i {
		return nil, fmt.Errorf("invalid listen fd index: %d", i)
	}
	file := files[i]

	var fdNetwork string
	switch network {
	case "sd":
		fdNetwork = "fd"
	case "sdgram":
		fdNetwork = "fdgram"
	default:
		return nil, fmt.Errorf("invalid network: %s", network)
	}

	na, err := caddy.ParseNetworkAddress(caddy.JoinNetworkAddress(fdNetwork, strconv.Itoa(file), port))
	if err != nil {
		return nil, err
	}

	return na.Listen(ctx, 0, cfg)
}
