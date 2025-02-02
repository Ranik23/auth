package sendnotificationservice

import (
	pb "auth/proto/notification"
)


type SendNotificationService struct {
	pb.UnimplementedNotificationServiceServer
}

func (s *SendNotificationService) SendNotification(req *pb.SendNotificationRequest) (*pb.SendNotificationResponse, error) {
	return nil, nil
}