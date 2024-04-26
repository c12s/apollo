package client

import (
	"context"
	//"apollo/proto1"
	"fmt"
	"log"
	oort "github.com/c12s/oort/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func AuthorizeUser(permission string, subjectId string) bool {
	conn, err := grpc.Dial("oort:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	evaluatorClient := oort.NewOortEvaluatorClient(conn)

	getResp, err := evaluatorClient.Authorize(context.Background(), &oort.AuthorizationReq{
			Subject:        &oort.Resource{
				Id:   subjectId,
				Kind: "user",
			},
			Object:         &oort.Resource{
				Id:   "idk",
				Kind: "user",
			},
			PermissionName: permission,
	}) 
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(getResp.Authorized)
	}

	return getResp.Authorized
}

func CreateOrgUserRelationship(org_id string, user_id string) error {
	conn, err := grpc.Dial("oort:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	administratorClient := oort.NewOortAdministratorClient(conn)

	log.Printf("Org za inherit: " + org_id)
	log.Printf("User za inherit: " + user_id)
	_, err = administratorClient.CreateInheritanceRel(context.TODO(), &oort.CreateInheritanceRelReq{
		From: &oort.Resource{
			Id:   org_id,
			Kind: "org",
		},
		To:   &oort.Resource{
			Id:   user_id,
			Kind: "user",
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func GetGrantedPermissions(user string) []*oort.GrantedPermission {
	conn, err := grpc.Dial("oort:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	evaluatorClient := oort.NewOortEvaluatorClient(conn)

	resp, err := evaluatorClient.GetGrantedPermissions(context.TODO(), &oort.GetGrantedPermissionsReq{
		Subject: &oort.Resource{
			Id:   user,
			Kind: "user",
		},
	})

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("permissions of user")
	for _, perm := range resp.Permissions {
		log.Printf("%s - %s/%s", perm.Name, perm.Object.Kind, perm.Object.Id)
	}

	return resp.Permissions
}

func CreatePolicyAsync(org_id string, user string, permissions []*oort.Permission) {
	administratorAsync, err := oort.NewAdministrationAsyncClient("nats:4222")

	if err != nil {
		log.Printf("Error calling CreatePolicyAsync: %s", err)
	}

	log.Printf("User za policy: " + user)
	log.Printf("Org za policy: " + org_id)
	for _, perm := range permissions {
		err := administratorAsync.SendRequest(&oort.CreatePolicyReq{
			SubjectScope: &oort.Resource{
				Id:   user,
				Kind: "user",
			},
			ObjectScope:  &oort.Resource{
				Id:   org_id,
				Kind: "org",
			},
			Permission:   perm,
		}, func(resp *oort.AdministrationAsyncResp) {
			if len(resp.Error) > 0 {
				log.Println(resp.Error)
			}
		})
		if err != nil {
			log.Fatalln(err)
		}
	}
	
}
