package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/cloudstax/openmanage/common"
	"github.com/cloudstax/openmanage/utils"
)

const (
	// Service members need to be created in advance. So TaskID, ContainerInstanceID
	// and ServerInstanceID would be empty at service member creation.
	// set them to default values, this will help the later conditional update.
	DefaultTaskID              = "defaultTaskID"
	DefaultContainerInstanceID = "defaultContainerInstanceID"
	DefaultServerInstanceID    = "defaultServerInstanceID"
)

func CreateDevice(cluster string, device string, service string) *common.Device {
	return &common.Device{
		ClusterName: cluster,
		DeviceName:  device,
		ServiceName: service,
	}
}

func EqualDevice(t1 *common.Device, t2 *common.Device) bool {
	if t1.ClusterName == t2.ClusterName &&
		t1.DeviceName == t2.DeviceName &&
		t1.ServiceName == t2.ServiceName {
		return true
	}
	return false
}

func CreateService(cluster string, service string, serviceUUID string) *common.Service {
	return &common.Service{
		ClusterName: cluster,
		ServiceName: service,
		ServiceUUID: serviceUUID,
	}
}

func EqualService(t1 *common.Service, t2 *common.Service) bool {
	if t1.ClusterName == t2.ClusterName &&
		t1.ServiceName == t2.ServiceName &&
		t1.ServiceUUID == t2.ServiceUUID {
		return true
	}
	return false
}

func CreateInitialServiceAttr(serviceUUID string, replicas int64, volSizeGB int64,
	cluster string, service string, devName string,
	registerDNS bool, domain string, hostedZoneID string) *common.ServiceAttr {
	return &common.ServiceAttr{
		ServiceUUID:   serviceUUID,
		ServiceStatus: common.ServiceStatusCreating,
		LastModified:  time.Now().UnixNano(),
		Replicas:      replicas,
		VolumeSizeGB:  volSizeGB,
		ClusterName:   cluster,
		ServiceName:   service,
		DeviceName:    devName,
		RegisterDNS:   registerDNS,
		DomainName:    domain,
		HostedZoneID:  hostedZoneID,
	}
}

func CreateServiceAttr(serviceUUID string, status string, mtime int64, replicas int64, volSizeGB int64,
	cluster string, service string, devName string,
	registerDNS bool, domain string, hostedZoneID string) *common.ServiceAttr {
	return &common.ServiceAttr{
		ServiceUUID:   serviceUUID,
		ServiceStatus: status,
		LastModified:  mtime,
		Replicas:      replicas,
		VolumeSizeGB:  volSizeGB,
		ClusterName:   cluster,
		ServiceName:   service,
		DeviceName:    devName,
		RegisterDNS:   registerDNS,
		DomainName:    domain,
		HostedZoneID:  hostedZoneID,
	}
}

func EqualServiceAttr(t1 *common.ServiceAttr, t2 *common.ServiceAttr, skipMtime bool) bool {
	if t1.ServiceUUID == t2.ServiceUUID &&
		t1.ServiceStatus == t2.ServiceStatus &&
		(skipMtime || t1.LastModified == t2.LastModified) &&
		t1.Replicas == t2.Replicas &&
		t1.VolumeSizeGB == t2.VolumeSizeGB &&
		t1.ClusterName == t2.ClusterName &&
		t1.ServiceName == t2.ServiceName &&
		t1.DeviceName == t2.DeviceName &&
		t1.RegisterDNS == t2.RegisterDNS &&
		t1.DomainName == t2.DomainName &&
		t1.HostedZoneID == t2.HostedZoneID {
		return true
	}
	return false
}

func UpdateServiceAttr(t1 *common.ServiceAttr, status string) *common.ServiceAttr {
	return &common.ServiceAttr{
		ServiceUUID:   t1.ServiceUUID,
		ServiceStatus: status,
		LastModified:  time.Now().UnixNano(),
		Replicas:      t1.Replicas,
		VolumeSizeGB:  t1.VolumeSizeGB,
		ClusterName:   t1.ClusterName,
		ServiceName:   t1.ServiceName,
		DeviceName:    t1.DeviceName,
		RegisterDNS:   t1.RegisterDNS,
		DomainName:    t1.DomainName,
		HostedZoneID:  t1.HostedZoneID,
	}
}

func CreateInitialServiceMember(serviceUUID string, volID string, devName string, az string,
	memberName string, configs []*common.MemberConfig) *common.ServiceMember {
	return &common.ServiceMember{
		ServiceUUID:         serviceUUID,
		VolumeID:            volID,
		LastModified:        time.Now().UnixNano(),
		DeviceName:          devName,
		AvailableZone:       az,
		TaskID:              DefaultTaskID,
		ContainerInstanceID: DefaultContainerInstanceID,
		ServerInstanceID:    DefaultServerInstanceID,
		MemberName:          memberName,
		Configs:             configs,
	}
}

func CreateServiceMember(serviceUUID string, volID string, mtime int64, devName string, az string,
	taskID string, containerInstanceID string, ec2InstanceID string, memberName string, configs []*common.MemberConfig) *common.ServiceMember {
	return &common.ServiceMember{
		ServiceUUID:         serviceUUID,
		VolumeID:            volID,
		LastModified:        mtime,
		DeviceName:          devName,
		AvailableZone:       az,
		TaskID:              taskID,
		ContainerInstanceID: containerInstanceID,
		ServerInstanceID:    ec2InstanceID,
		MemberName:          memberName,
		Configs:             configs,
	}
}

func EqualServiceMember(t1 *common.ServiceMember, t2 *common.ServiceMember, skipMtime bool) bool {
	if t1.ServiceUUID == t2.ServiceUUID &&
		t1.VolumeID == t2.VolumeID &&
		(skipMtime || t1.LastModified == t2.LastModified) &&
		t1.DeviceName == t2.DeviceName &&
		t1.AvailableZone == t2.AvailableZone &&
		t1.TaskID == t2.TaskID &&
		t1.ContainerInstanceID == t2.ContainerInstanceID &&
		t1.ServerInstanceID == t2.ServerInstanceID &&
		t1.MemberName == t2.MemberName &&
		equalConfigs(t1.Configs, t2.Configs) {
		return true
	}
	return false
}

func equalConfigs(c1 []*common.MemberConfig, c2 []*common.MemberConfig) bool {
	if len(c1) != len(c2) {
		return false
	}
	for i := 0; i < len(c1); i++ {
		if c1[i].FileName != c2[i].FileName ||
			c1[i].FileID != c2[i].FileID ||
			c1[i].FileMD5 != c2[i].FileMD5 {
			return false
		}
	}
	return true
}

func CopyMemberConfigs(c1 []*common.MemberConfig) []*common.MemberConfig {
	c2 := make([]*common.MemberConfig, len(c1))
	for i, c := range c1 {
		c2[i] = &common.MemberConfig{
			FileName: c.FileName,
			FileID:   c.FileID,
			FileMD5:  c.FileMD5,
		}
	}
	return c2
}

func UpdateServiceMemberConfigs(t1 *common.ServiceMember, c []*common.MemberConfig) *common.ServiceMember {
	return &common.ServiceMember{
		ServiceUUID:         t1.ServiceUUID,
		VolumeID:            t1.VolumeID,
		LastModified:        time.Now().UnixNano(),
		DeviceName:          t1.DeviceName,
		AvailableZone:       t1.AvailableZone,
		TaskID:              t1.TaskID,
		ContainerInstanceID: t1.ContainerInstanceID,
		ServerInstanceID:    t1.ServerInstanceID,
		MemberName:          t1.MemberName,
		Configs:             c,
	}
}

func UpdateServiceMemberOwner(t1 *common.ServiceMember, taskID string, containerInstanceID string, ec2InstanceID string) *common.ServiceMember {
	return &common.ServiceMember{
		ServiceUUID:         t1.ServiceUUID,
		VolumeID:            t1.VolumeID,
		LastModified:        time.Now().UnixNano(),
		DeviceName:          t1.DeviceName,
		AvailableZone:       t1.AvailableZone,
		TaskID:              taskID,
		ContainerInstanceID: containerInstanceID,
		ServerInstanceID:    ec2InstanceID,
		MemberName:          t1.MemberName,
		Configs:             t1.Configs,
	}
}

func CreateInitialConfigFile(serviceUUID string, fileID string, fileName string, fileMode uint32, content string) *common.ConfigFile {
	chksum := utils.GenMD5(content)
	return &common.ConfigFile{
		ServiceUUID:  serviceUUID,
		FileID:       fileID,
		FileMD5:      chksum,
		FileName:     fileName,
		FileMode:     fileMode,
		LastModified: time.Now().UnixNano(),
		Content:      content,
	}
}

func CreateConfigFile(serviceUUID string, fileID string, fileMD5 string,
	fileName string, fileMode uint32, mtime int64, content string) (*common.ConfigFile, error) {
	// double check config file
	chksum := utils.GenMD5(content)
	if chksum != fileMD5 {
		errmsg := fmt.Sprintf("internal error, file %s content corrupted, expect md5 %s content md5 %s",
			fileID, fileMD5, chksum)
		return nil, errors.New(errmsg)
	}

	cfg := &common.ConfigFile{
		ServiceUUID:  serviceUUID,
		FileID:       fileID,
		FileMD5:      fileMD5,
		FileName:     fileName,
		FileMode:     fileMode,
		LastModified: mtime,
		Content:      content,
	}
	return cfg, nil
}

func UpdateConfigFile(c *common.ConfigFile, newFileID string, newContent string) *common.ConfigFile {
	newMD5 := utils.GenMD5(newContent)
	return &common.ConfigFile{
		ServiceUUID:  c.ServiceUUID,
		FileID:       newFileID,
		FileMD5:      newMD5,
		FileName:     c.FileName,
		FileMode:     c.FileMode,
		LastModified: time.Now().UnixNano(),
		Content:      newContent,
	}
}

func EqualConfigFile(c1 *common.ConfigFile, c2 *common.ConfigFile, skipMtime bool, skipContent bool) bool {
	if c1.ServiceUUID == c2.ServiceUUID &&
		c1.FileID == c2.FileID &&
		c1.FileMD5 == c2.FileMD5 &&
		c1.FileName == c2.FileName &&
		c1.FileMode == c2.FileMode &&
		(skipMtime || c1.LastModified == c2.LastModified) &&
		(skipContent || c1.Content == c2.Content) {
		return true
	}
	return false
}

func PrintConfigFile(cfg *common.ConfigFile) string {
	return fmt.Sprintf("serviceUUID %s fileID %s fileName %s fileMD5 %s fileMode %d LastModified %d",
		cfg.ServiceUUID, cfg.FileID, cfg.FileName, cfg.FileMD5, cfg.FileMode, cfg.LastModified)
}
