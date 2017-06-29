package db

import (
	"sync"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/cloudstax/openmanage/common"
)

type MemDB struct {
	devMap     map[string]common.Device
	svcMap     map[string]common.Service
	svcAttrMap map[string]common.ServiceAttr
	volMap     map[string]common.Volume
	cfgMap     map[string]common.ConfigFile
	mlock      *sync.Mutex
}

func NewMemDB() *MemDB {
	d := &MemDB{
		devMap:     map[string]common.Device{},
		svcMap:     map[string]common.Service{},
		svcAttrMap: map[string]common.ServiceAttr{},
		volMap:     map[string]common.Volume{},
		cfgMap:     map[string]common.ConfigFile{},
		mlock:      &sync.Mutex{},
	}
	return d
}

func (d *MemDB) CreateSystemTables(ctx context.Context) error {
	return nil
}

func (d *MemDB) SystemTablesReady(ctx context.Context) (tableStatus string, ready bool, err error) {
	return TableStatusActive, true, nil
}

func (d *MemDB) DeleteSystemTables(ctx context.Context) error {
	return nil
}

func (d *MemDB) CreateDevice(ctx context.Context, dev *common.Device) error {
	key := dev.ClusterName + dev.DeviceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.devMap[key]
	if ok {
		glog.Errorln("device exists", key)
		return ErrDBConditionalCheckFailed
	}

	d.devMap[key] = *dev
	return nil
}

func (d *MemDB) GetDevice(ctx context.Context, clusterName string, deviceName string) (dev *common.Device, err error) {
	key := clusterName + deviceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	cdev, ok := d.devMap[key]
	if !ok {
		glog.Errorln("device not found", key)
		return nil, ErrDBRecordNotFound
	}
	return copyDevice(&cdev), nil
}

func (d *MemDB) DeleteDevice(ctx context.Context, clusterName string, deviceName string) error {
	key := clusterName + deviceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.devMap[key]
	if !ok {
		glog.Errorln("device not exist", key)
		return ErrDBRecordNotFound
	}

	delete(d.devMap, key)
	return nil
}

func (d *MemDB) ListDevices(ctx context.Context, clusterName string) (devs []*common.Device, err error) {
	return d.listDevicesWithLimit(ctx, clusterName, 0)
}

func (d *MemDB) listDevicesWithLimit(ctx context.Context, clusterName string, limit int64) (devs []*common.Device, err error) {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	devs = make([]*common.Device, len(d.devMap))
	idx := 0
	for _, dev := range d.devMap {
		devs[idx] = copyDevice(&dev)
		idx++
	}
	return devs, nil
}

func (d *MemDB) CreateService(ctx context.Context, svc *common.Service) error {
	key := svc.ClusterName + svc.ServiceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.svcMap[key]
	if ok {
		glog.Errorln("service exists", key)
		return ErrDBConditionalCheckFailed
	}

	d.svcMap[key] = *svc
	return nil
}

func (d *MemDB) GetService(ctx context.Context, clusterName string, serviceName string) (svc *common.Service, err error) {
	key := clusterName + serviceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	csvc, ok := d.svcMap[key]
	if !ok {
		glog.Errorln("service not exist", key)
		return nil, ErrDBRecordNotFound
	}
	return copyService(&csvc), nil
}

func (d *MemDB) DeleteService(ctx context.Context, clusterName string, serviceName string) error {
	key := clusterName + serviceName

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.svcMap[key]
	if !ok {
		glog.Errorln("service not exist", key)
		return ErrDBRecordNotFound
	}

	delete(d.svcMap, key)
	return nil
}

func (d *MemDB) ListServices(ctx context.Context, clusterName string) (svcs []*common.Service, err error) {
	return d.listServicesWithLimit(ctx, clusterName, 0)
}

func (d *MemDB) listServicesWithLimit(ctx context.Context, clusterName string, limit int64) (svcs []*common.Service, err error) {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	svcs = make([]*common.Service, len(d.svcMap))
	idx := 0
	for _, svc := range d.svcMap {
		svcs[idx] = copyService(&svc)
		idx++
	}
	return svcs, nil
}

func (d *MemDB) CreateServiceAttr(ctx context.Context, attr *common.ServiceAttr) error {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.svcAttrMap[attr.ServiceUUID]
	if ok {
		glog.Errorln("ServiceAttr exists", attr)
		return ErrDBConditionalCheckFailed
	}

	d.svcAttrMap[attr.ServiceUUID] = *attr
	return nil
}

func (d *MemDB) UpdateServiceAttr(ctx context.Context, oldAttr *common.ServiceAttr, newAttr *common.ServiceAttr) error {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.svcAttrMap[oldAttr.ServiceUUID]
	if !ok {
		glog.Errorln("serviceAttr not exist", oldAttr)
		return ErrDBRecordNotFound
	}

	d.svcAttrMap[oldAttr.ServiceUUID] = *newAttr
	return nil
}

func (d *MemDB) GetServiceAttr(ctx context.Context, serviceUUID string) (attr *common.ServiceAttr, err error) {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	cattr, ok := d.svcAttrMap[serviceUUID]
	if !ok {
		glog.Errorln("ServiceAttr not exists", serviceUUID)
		return nil, ErrDBRecordNotFound
	}

	return copyServiceAttr(&cattr), nil
}

func (d *MemDB) DeleteServiceAttr(ctx context.Context, serviceUUID string) error {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.svcAttrMap[serviceUUID]
	if !ok {
		glog.Errorln("ServiceAttr not exists", serviceUUID)
		return ErrDBRecordNotFound
	}

	delete(d.svcAttrMap, serviceUUID)
	return nil
}

func (d *MemDB) CreateVolume(ctx context.Context, vol *common.Volume) error {
	key := vol.ServiceUUID + vol.VolumeID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.volMap[key]
	if ok {
		glog.Errorln("volume exists", key)
		return ErrDBConditionalCheckFailed
	}

	d.volMap[key] = *vol
	return nil
}

func (d *MemDB) UpdateVolume(ctx context.Context, oldVol *common.Volume, newVol *common.Volume) error {
	key := oldVol.ServiceUUID + oldVol.VolumeID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	vol, ok := d.volMap[key]
	if !ok {
		glog.Errorln("volume not exist", key)
		return ErrDBRecordNotFound
	}
	if !EqualVolume(oldVol, &vol, true) {
		glog.Errorln("oldVol not match current vol, oldVol", oldVol, "current vol", vol)
		return ErrDBConditionalCheckFailed
	}

	d.volMap[key] = *newVol
	return nil
}

func (d *MemDB) GetVolume(ctx context.Context, serviceUUID string, volumeID string) (vol *common.Volume, err error) {
	key := serviceUUID + volumeID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	cvol, ok := d.volMap[key]
	if !ok {
		glog.Errorln("volume not exist", key)
		return nil, ErrDBRecordNotFound
	}

	return copyVolume(&cvol), nil
}

func (d *MemDB) ListVolumes(ctx context.Context, serviceUUID string) (vols []*common.Volume, err error) {
	return d.listVolumesWithLimit(ctx, serviceUUID, 0)
}

func (d *MemDB) listVolumesWithLimit(ctx context.Context, serviceUUID string, limit int64) (vols []*common.Volume, err error) {
	d.mlock.Lock()
	defer d.mlock.Unlock()

	for _, vol := range d.volMap {
		if vol.ServiceUUID == serviceUUID {
			vols = append(vols, copyVolume(&vol))
		}
	}
	return vols, nil
}

func (d *MemDB) DeleteVolume(ctx context.Context, serviceUUID string, volumeID string) error {
	key := serviceUUID + volumeID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.volMap[key]
	if !ok {
		glog.Errorln("volume not exist", key)
		return ErrDBRecordNotFound
	}

	delete(d.volMap, key)
	return nil
}

func (d *MemDB) CreateConfigFile(ctx context.Context, cfg *common.ConfigFile) error {
	key := cfg.ServiceUUID + cfg.FileID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.cfgMap[key]
	if ok {
		glog.Errorln("config file exists", key)
		return ErrDBConditionalCheckFailed
	}

	d.cfgMap[key] = *cfg
	return nil
}

func (d *MemDB) GetConfigFile(ctx context.Context, serviceUUID string, fileID string) (cfg *common.ConfigFile, err error) {
	key := serviceUUID + fileID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	c, ok := d.cfgMap[key]
	if !ok {
		glog.Errorln("config file not exist", key)
		return nil, ErrDBRecordNotFound
	}

	return copyConfigFile(&c), nil
}

func (d *MemDB) DeleteConfigFile(ctx context.Context, serviceUUID string, fileID string) error {
	key := serviceUUID + fileID

	d.mlock.Lock()
	defer d.mlock.Unlock()

	_, ok := d.cfgMap[key]
	if !ok {
		glog.Errorln("config file not exist", key)
		return ErrDBRecordNotFound
	}

	delete(d.cfgMap, key)
	return nil
}

func copyDevice(t *common.Device) *common.Device {
	return &common.Device{
		ClusterName: t.ClusterName,
		DeviceName:  t.DeviceName,
		ServiceName: t.ServiceName,
	}
}

func copyService(t *common.Service) *common.Service {
	return &common.Service{
		ClusterName: t.ClusterName,
		ServiceName: t.ServiceName,
		ServiceUUID: t.ServiceUUID,
	}
}

func copyServiceAttr(t *common.ServiceAttr) *common.ServiceAttr {
	return &common.ServiceAttr{
		ServiceUUID:         t.ServiceUUID,
		ServiceStatus:       t.ServiceStatus,
		LastModified:        t.LastModified,
		Replicas:            t.Replicas,
		VolumeSizeGB:        t.VolumeSizeGB,
		ClusterName:         t.ClusterName,
		ServiceName:         t.ServiceName,
		DeviceName:          t.DeviceName,
		HasStrictMembership: t.HasStrictMembership,
		DomainName:          t.DomainName,
		HostedZoneID:        t.HostedZoneID,
	}
}

func copyVolume(t *common.Volume) *common.Volume {
	return &common.Volume{
		ServiceUUID:         t.ServiceUUID,
		VolumeID:            t.VolumeID,
		LastModified:        t.LastModified,
		DeviceName:          t.DeviceName,
		AvailableZone:       t.AvailableZone,
		TaskID:              t.TaskID,
		ContainerInstanceID: t.ContainerInstanceID,
		ServerInstanceID:    t.ServerInstanceID,
		MemberName:          t.MemberName,
		Configs:             t.Configs,
	}
}

func copyConfigFile(t *common.ConfigFile) *common.ConfigFile {
	return &common.ConfigFile{
		FileID:       t.FileID,
		FileMD5:      t.FileMD5,
		FileName:     t.FileName,
		FileMode:     t.FileMode,
		ServiceUUID:  t.ServiceUUID,
		LastModified: t.LastModified,
		Content:      t.Content,
	}
}
