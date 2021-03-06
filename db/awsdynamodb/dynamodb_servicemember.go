package awsdynamodb

import (
	"encoding/json"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/cloudstax/openmanage/common"
	"github.com/cloudstax/openmanage/db"
	"github.com/cloudstax/openmanage/utils"
)

// CreateServiceMember creates one EBS serviceMember in DB
func (d *DynamoDB) CreateServiceMember(ctx context.Context, member *common.ServiceMember) error {
	requuid := utils.GetReqIDFromContext(ctx)
	configBytes, err := json.Marshal(member.Configs)
	if err != nil {
		glog.Errorln("Marshal MemberConfigs error", err, member, "requuid", requuid)
		return err
	}

	dbsvc := dynamodb.New(d.sess)

	params := &dynamodb.PutItemInput{
		TableName: aws.String(d.serviceMemberTableName),
		Item: map[string]*dynamodb.AttributeValue{
			db.ServiceUUID: {
				S: aws.String(member.ServiceUUID),
			},
			db.MemberName: {
				S: aws.String(member.MemberName),
			},
			db.LastModified: {
				N: aws.String(strconv.FormatInt(member.LastModified, 10)),
			},
			db.AvailableZone: {
				S: aws.String(member.AvailableZone),
			},
			db.TaskID: {
				S: aws.String(member.TaskID),
			},
			db.ContainerInstanceID: {
				S: aws.String(member.ContainerInstanceID),
			},
			db.ServerInstanceID: {
				S: aws.String(member.ServerInstanceID),
			},
			db.VolumeID: {
				S: aws.String(member.VolumeID),
			},
			db.DeviceName: {
				S: aws.String(member.DeviceName),
			},
			db.MemberConfigs: {
				B: configBytes,
			},
		},
		ConditionExpression:    aws.String(db.ServiceMemberPutCondition),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
	resp, err := dbsvc.PutItem(params)

	if err != nil {
		glog.Errorln("failed to create serviceMember", member, "error", err, "requuid", requuid)
		return d.convertError(err)
	}

	glog.Infoln("created serviceMember", member, "requuid", requuid, "resp", resp)
	return nil
}

// UpdateServiceMember updates the db.ServiceMember in DB
func (d *DynamoDB) UpdateServiceMember(ctx context.Context, oldMember *common.ServiceMember, newMember *common.ServiceMember) error {
	requuid := utils.GetReqIDFromContext(ctx)

	// sanity check. ServiceUUID, VolumeID, etc, are not allowed to update.
	if oldMember.ServiceUUID != newMember.ServiceUUID ||
		oldMember.VolumeID != newMember.VolumeID ||
		oldMember.DeviceName != newMember.DeviceName ||
		oldMember.AvailableZone != newMember.AvailableZone ||
		oldMember.MemberName != newMember.MemberName {
		glog.Errorln("immutable attributes are updated, oldMember", oldMember, "newMember", newMember, "requuid", requuid)
		return db.ErrDBInvalidRequest
	}

	var err error
	var oldCfgBytes []byte
	if oldMember.Configs != nil {
		oldCfgBytes, err = json.Marshal(oldMember.Configs)
		if err != nil {
			glog.Errorln("Marshal new MemberConfigs error", err, "requuid", requuid, oldMember.Configs)
			return err
		}
	}

	var newCfgBytes []byte
	if newMember.Configs != nil {
		newCfgBytes, err = json.Marshal(newMember.Configs)
		if err != nil {
			glog.Errorln("Marshal new MemberConfigs error", err, "requuid", requuid, newMember.Configs)
			return err
		}
	}

	dbsvc := dynamodb.New(d.sess)

	updateExpr := "SET " + db.TaskID + " = :v1, " + db.ContainerInstanceID + " = :v2, " +
		db.ServerInstanceID + " = :v3, " + db.LastModified + " = :v4, " + db.MemberConfigs + " = :v5"
	conditionExpr := db.TaskID + " = :cv1 AND " + db.ContainerInstanceID + " = :cv2 AND " +
		db.ServerInstanceID + " = :cv3 AND " + db.MemberConfigs + " = :cv4"

	params := &dynamodb.UpdateItemInput{
		TableName: aws.String(d.serviceMemberTableName),
		Key: map[string]*dynamodb.AttributeValue{
			db.ServiceUUID: {
				S: aws.String(oldMember.ServiceUUID),
			},
			db.MemberName: {
				S: aws.String(oldMember.MemberName),
			},
		},
		UpdateExpression:    aws.String(updateExpr),
		ConditionExpression: aws.String(conditionExpr),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":v1": {
				S: aws.String(newMember.TaskID),
			},
			":v2": {
				S: aws.String(newMember.ContainerInstanceID),
			},
			":v3": {
				S: aws.String(newMember.ServerInstanceID),
			},
			":v4": {
				N: aws.String(strconv.FormatInt(newMember.LastModified, 10)),
			},
			":v5": {
				B: newCfgBytes,
			},
			":cv1": {
				S: aws.String(oldMember.TaskID),
			},
			":cv2": {
				S: aws.String(oldMember.ContainerInstanceID),
			},
			":cv3": {
				S: aws.String(oldMember.ServerInstanceID),
			},
			":cv4": {
				B: oldCfgBytes,
			},
		},
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}

	resp, err := dbsvc.UpdateItem(params)

	if err != nil {
		glog.Errorln("failed to update serviceMember", oldMember, "to", newMember, "error", err, "requuid", requuid)
		return d.convertError(err)
	}

	glog.Infoln("updated serviceMember", oldMember, "to", newMember, "requuid", requuid, "resp", resp)
	return nil
}

// GetServiceMember gets the serviceMemberItem from DB
func (d *DynamoDB) GetServiceMember(ctx context.Context, serviceUUID string, memberName string) (member *common.ServiceMember, err error) {
	requuid := utils.GetReqIDFromContext(ctx)
	dbsvc := dynamodb.New(d.sess)

	params := &dynamodb.GetItemInput{
		TableName: aws.String(d.serviceMemberTableName),
		Key: map[string]*dynamodb.AttributeValue{
			db.ServiceUUID: {
				S: aws.String(serviceUUID),
			},
			db.MemberName: {
				S: aws.String(memberName),
			},
		},
		ConsistentRead:         aws.Bool(true),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}
	resp, err := dbsvc.GetItem(params)

	if err != nil {
		glog.Errorln("failed to get serviceMember", memberName, "serviceUUID", serviceUUID, "error", err, "requuid", requuid)
		return nil, d.convertError(err)
	}

	if len(resp.Item) == 0 {
		glog.Infoln("serviceMember", memberName, "not found, serviceUUID", serviceUUID, "requuid", requuid)
		return nil, db.ErrDBRecordNotFound
	}

	member, err = d.attrsToServiceMember(resp.Item)
	if err != nil {
		glog.Errorln("GetServiceMember convert dynamodb attributes to serviceMember error", err, "requuid", requuid, "resp", resp)
		return nil, err
	}

	glog.Infoln("get serviceMember", member, "requuid", requuid)
	return member, nil
}

// ListServiceMembers lists all serviceMembers of the service
func (d *DynamoDB) ListServiceMembers(ctx context.Context, serviceUUID string) (serviceMembers []*common.ServiceMember, err error) {
	return d.listServiceMembersWithLimit(ctx, serviceUUID, 0)
}

// listServiceMembersWithLimit limits the returned db.ServiceMembers at one query.
// This is for testing the pagination list.
func (d *DynamoDB) listServiceMembersWithLimit(ctx context.Context, serviceUUID string, limit int64) (serviceMembers []*common.ServiceMember, err error) {
	requuid := utils.GetReqIDFromContext(ctx)
	dbsvc := dynamodb.New(d.sess)

	var lastEvaluatedKey map[string]*dynamodb.AttributeValue
	lastEvaluatedKey = nil

	for true {
		params := &dynamodb.QueryInput{
			TableName:              aws.String(d.serviceMemberTableName),
			KeyConditionExpression: aws.String(db.ServiceUUID + " = :v1"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":v1": {
					S: aws.String(serviceUUID),
				},
			},
			ConsistentRead:         aws.Bool(true),
			ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
		}
		if limit > 0 {
			// set the query limit
			params.Limit = aws.Int64(limit)
		}
		if len(lastEvaluatedKey) != 0 {
			params.ExclusiveStartKey = lastEvaluatedKey
		}

		resp, err := dbsvc.Query(params)

		if err != nil {
			glog.Errorln("failed to list serviceMembers, serviceUUID", serviceUUID,
				"limit", limit, "lastEvaluatedKey", lastEvaluatedKey, "error", err, "requuid", requuid)
			return nil, d.convertError(err)
		}

		glog.Infoln("list serviceMembers succeeded, serviceUUID",
			serviceUUID, "limit", limit, "requuid", requuid, "resp count", resp.Count)

		lastEvaluatedKey = resp.LastEvaluatedKey

		if len(resp.Items) == 0 {
			// is it possible dynamodb returns no items with LastEvaluatedKey?
			// we don't use complex filter, so would be impossible?
			if len(resp.LastEvaluatedKey) != 0 {
				glog.Errorln("no items in resp but LastEvaluatedKey is not empty, resp", resp, "requuid", requuid)
				continue
			}

			glog.Infoln("no more serviceMember item for serviceUUID",
				serviceUUID, "serviceMembers", len(serviceMembers), "requuid", requuid)
			return serviceMembers, nil
		}

		for _, item := range resp.Items {
			member, err := d.attrsToServiceMember(item)
			if err != nil {
				glog.Errorln("ListServiceMember convert dynamodb attributes to serviceMember error", err, "requuid", requuid, "item", item)
				return nil, err
			}
			serviceMembers = append(serviceMembers, member)
		}

		glog.Infoln("list", len(serviceMembers), "serviceMembers, serviceUUID",
			serviceUUID, "LastEvaluatedKey", lastEvaluatedKey, "requuid", requuid)

		if len(lastEvaluatedKey) == 0 {
			// no more serviceMembers
			break
		}
	}

	return serviceMembers, nil
}

// DeleteServiceMember deletes the serviceMember from DB
func (d *DynamoDB) DeleteServiceMember(ctx context.Context, serviceUUID string, memberName string) error {
	requuid := utils.GetReqIDFromContext(ctx)
	dbsvc := dynamodb.New(d.sess)

	// TODO reject if any serviceMember is still attached or service item is not at DELETING status.

	params := &dynamodb.DeleteItemInput{
		TableName: aws.String(d.serviceMemberTableName),
		Key: map[string]*dynamodb.AttributeValue{
			db.ServiceUUID: {
				S: aws.String(serviceUUID),
			},
			db.MemberName: {
				S: aws.String(memberName),
			},
		},
		ConditionExpression:    aws.String(db.ServiceMemberDelCondition),
		ReturnConsumedCapacity: aws.String(dynamodb.ReturnConsumedCapacityTotal),
	}

	resp, err := dbsvc.DeleteItem(params)

	if err != nil {
		if err.(awserr.Error).Code() == ConditionalCheckFailedException {
			glog.Infoln("serviceMember not found", memberName, "serviceUUID", serviceUUID, "requuid", requuid, "resp", resp)
			return db.ErrDBRecordNotFound
		}
		glog.Errorln("failed to delete serviceMember", memberName,
			"serviceUUID", serviceUUID, "error", err, "requuid", requuid)
		return d.convertError(err)
	}

	glog.Infoln("deleted serviceMember", memberName, "serviceUUID", serviceUUID, "requuid", requuid, "resp", resp)
	return nil
}

func (d *DynamoDB) attrsToServiceMember(item map[string]*dynamodb.AttributeValue) (*common.ServiceMember, error) {
	mtime, err := strconv.ParseInt(*(item[db.LastModified].N), 10, 64)
	if err != nil {
		glog.Errorln("ParseInt LastModified error", err, item)
		return nil, db.ErrDBInternal
	}

	var configs []*common.MemberConfig
	err = json.Unmarshal(item[db.MemberConfigs].B, &configs)
	if err != nil {
		glog.Errorln("Unmarshal json MemberConfigs error", err, item)
		return nil, db.ErrDBInternal
	}

	member := db.CreateServiceMember(*(item[db.ServiceUUID].S),
		*(item[db.VolumeID].S),
		mtime,
		*(item[db.DeviceName].S),
		*(item[db.AvailableZone].S),
		*(item[db.TaskID].S),
		*(item[db.ContainerInstanceID].S),
		*(item[db.ServerInstanceID].S),
		*(item[db.MemberName].S),
		configs)

	return member, nil
}
