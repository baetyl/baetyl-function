package resolve

import (
	"fmt"
	"sync"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/native"
	"github.com/fsnotify/fsnotify"
)

type NativeResolver struct {
	watcher *fsnotify.Watcher
	mapping *native.ServiceMapping
	log     *log.Logger
	sync.RWMutex
}

func NewNativeResolver(_ context.Context) (Resolver, error) {
	resolver := &NativeResolver{
		mapping: new(native.ServiceMapping),
		log:     log.With(log.Any("resolve", "native")),
	}

	err := resolver.mapping.Load()
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					resolver.log.Warn("error when wait on the events channel")
					return
				}
				resolver.log.Debug("received a file event", log.Any("eventName", event.Name), log.Any("eventOp", event.Op))

				if event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				resolver.log.Debug("load ports mapping file again", log.Error(err))
				err = resolver.mapping.Load()
				if err != nil {
					resolver.log.Warn("load ports mapping file failed", log.Error(err))
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					resolver.log.Warn("error when wait on the events channel")
					return
				}
				resolver.log.Warn("error when wait on the events channel", log.Error(err))
			}
		}
	}()

	err = watcher.Add(native.ServiceMappingFile)
	if err != nil {
		return nil, err
	}

	resolver.watcher = watcher
	return resolver, nil
}

func (n *NativeResolver) Resolve(service string) (address string, err error) {
	n.Lock()
	defer n.Unlock()

	serviceInfo, ok := n.mapping.Services[service]
	if !ok {
		return "", errors.New("no such service in services mapping file")
	}
	if len(serviceInfo.Ports.Items) == 0{
		return "", errors.New("no ports info in services mapping file")
	}

	port, err := serviceInfo.Ports.Next()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("localhost:%d", port), nil
}

func (n *NativeResolver) Close() error {
	err := n.watcher.Close()
	if err != nil {
		n.log.Warn("failed to close file watcher", log.Error(err))
	}
	return nil
}
