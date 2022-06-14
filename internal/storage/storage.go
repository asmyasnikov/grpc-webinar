package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"

	pbCDC "github.com/amasynikov/grpc-webinar/internal/genproto/cdc"
	pbCRUD "github.com/amasynikov/grpc-webinar/internal/genproto/crud"
)

type storageServer struct {
	pbCRUD.UnimplementedCRUDServer
	pbCDC.UnimplementedCDCServer

	// read-write access
	dataMtx sync.RWMutex
	data    map[string][]byte

	// read-write access
	listenersMtx sync.RWMutex
	listeners    map[pbCDC.CDC_ListenServer]chan struct{}

	cdcChannel chan *pbCDC.ListenResponse
}

func (c *storageServer) Listen(request *pbCDC.ListenRequest, listener pbCDC.CDC_ListenServer) error {
	c.listenersMtx.Lock()
	ch := make(chan struct{})
	c.listeners[listener] = ch
	c.listenersMtx.Unlock()
	<-ch
	return nil
}

func (c *storageServer) Create(ctx context.Context, request *pbCRUD.CreateRequest) (_ *pbCRUD.CreateResponse, err error) {
	log.Info().Caller().Msg("create")
	defer func() {
		if err != nil {
			log.Error().Caller().Msg("create failed")
		} else {
			log.Info().Caller().Msg("create done")
		}
	}()

	id, err := c.create(ctx, request.GetRaw())
	if err != nil {
		return nil, err
	}

	return &pbCRUD.CreateResponse{Id: id}, nil
}

func (c *storageServer) create(ctx context.Context, data []byte) (id string, err error) {
	defer func() {
		if err == nil {
			c.notify(pbCDC.ListenResponse_Created, &pbCDC.Data{
				Id:  id,
				Raw: data,
			})
		}
	}()

	uuid, err := uuid.NewUUID()
	if err != nil {
		return "", status.Errorf(codes.Internal, err.Error())
	}

	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	c.data[uuid.String()] = data

	return uuid.String(), nil
}

func (c *storageServer) Read(ctx context.Context, request *pbCRUD.ReadRequest) (_ *pbCRUD.ReadResponse, err error) {
	log.Info().Caller().Msg("read")
	defer func() {
		if err != nil {
			log.Error().Caller().Msg("read failed")
		} else {
			log.Info().Caller().Msg("read done")
		}
	}()

	data, err := c.read(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &pbCRUD.ReadResponse{Raw: data}, nil
}

func (c *storageServer) read(ctx context.Context, id string) (data []byte, err error) {
	c.dataMtx.RLock()
	defer c.dataMtx.RUnlock()

	if data, ok := c.data[id]; ok {
		return data, nil
	}

	return nil, status.Errorf(codes.NotFound, "")
}

func (c *storageServer) Update(ctx context.Context, request *pbCRUD.UpdateRequest) (_ *pbCRUD.UpdateResponse, err error) {
	log.Info().Caller().Msg("update")
	defer func() {
		if err != nil {
			log.Error().Caller().Msg("update failed")
		} else {
			log.Info().Caller().Msg("update done")
		}
	}()

	err = c.update(ctx, request.GetData().GetId(), request.GetData().GetRaw())
	if err != nil {
		return nil, err
	}

	return &pbCRUD.UpdateResponse{}, nil
}

func (c *storageServer) update(ctx context.Context, id string, data []byte) (err error) {
	defer func() {
		if err == nil {
			c.notify(pbCDC.ListenResponse_Updated, &pbCDC.Data{
				Id:  id,
				Raw: data,
			})
		}
	}()

	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	if _, ok := c.data[id]; !ok {
		return status.Errorf(codes.NotFound, "")
	}

	c.data[id] = data

	return nil
}

func (c *storageServer) Delete(ctx context.Context, request *pbCRUD.DeleteRequest) (_ *pbCRUD.DeleteResponse, err error) {
	log.Info().Caller().Msg("delete")
	defer func() {
		if err != nil {
			log.Error().Caller().Msg("delete failed")
		} else {
			log.Info().Caller().Msg("delete done")
		}
	}()

	err = c.delete(ctx, request.GetId())
	if err != nil {
		return nil, err
	}

	return &pbCRUD.DeleteResponse{}, nil
}

func (c *storageServer) delete(ctx context.Context, id string) (err error) {
	defer func() {
		if err == nil {
			c.notify(pbCDC.ListenResponse_Deleted, &pbCDC.Data{
				Id: id,
			})
		}
	}()

	c.dataMtx.Lock()
	defer c.dataMtx.Unlock()

	delete(c.data, id)

	return nil
}

func (c *storageServer) notify(event pbCDC.ListenResponse_EventType, data *pbCDC.Data) {
	c.cdcChannel <- &pbCDC.ListenResponse{
		Event: event,
		Data:  data,
	}
}

func (c *storageServer) sendChanges() {
	for msg := range c.cdcChannel {
		var listenersToDelete []pbCDC.CDC_ListenServer
		c.listenersMtx.RLock()
		for l, ch := range c.listeners {
			if err := l.Send(msg); err != nil {
				listenersToDelete = append(listenersToDelete, l)
				close(ch)
			}
		}
		c.listenersMtx.RUnlock()

		c.listenersMtx.Lock()
		for _, l := range listenersToDelete {
			delete(c.listeners, l)
		}
		c.listenersMtx.Unlock()
	}
}

func New() *storageServer {
	s := &storageServer{
		data:       make(map[string][]byte),
		listeners:  make(map[pbCDC.CDC_ListenServer]chan struct{}, 0),
		cdcChannel: make(chan *pbCDC.ListenResponse, 10),
	}
	go s.sendChanges()
	return s
}
