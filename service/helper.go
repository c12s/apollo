package service

import (
	"errors"
	"fmt"
	"github.com/c12s/apollo/helper"
	aPb "github.com/c12s/scheme/apollo"
	cPb "github.com/c12s/scheme/celestial"
	rPb "github.com/c12s/scheme/core"
	mPb "github.com/c12s/scheme/meridian"
	sg "github.com/c12s/stellar-go"
	"golang.org/x/net/context"
)

func (s *Server) auth(ctx context.Context, opt *aPb.AuthOpt) error {
	span, _ := sg.FromGRPCContext(ctx, "auth")
	defer span.Finish()
	fmt.Println(span)

	resp, err := s.Auth(ctx, opt)
	if err != nil {
		span.AddLog(&sg.KV{"apollo resp error", err.Error()})
		return err
	}

	if !resp.Value {
		span.AddLog(&sg.KV{"apollo.auth value", resp.Data["message"]})
		return errors.New(resp.Data["message"])
	}
	return nil
}

func (s *Server) checkNS(ctx context.Context, userid, namespace string) (string, error) {
	span, _ := sg.FromGRPCContext(ctx, "ns check")
	defer span.Finish()
	fmt.Println(span)

	client := NewMeridianClient(s.meridian)
	mrsp, err := client.Exists(sg.NewTracedGRPCContext(ctx, span),
		&mPb.NSReq{
			Name:   namespace,
			Extras: map[string]string{"userid": userid},
		},
	)
	if err != nil {
		span.AddLog(&sg.KV{"meridian exists error", err.Error()})
		return "", err
	}

	if mrsp.Extras["exists"] == "" {
		fmt.Println("namespace do not exists")
		return "", errors.New(fmt.Sprintf("%s do not exists", namespace))
	}
	fmt.Println("namespace exists")
	return mrsp.Extras["exists"], nil
}

func (s *Server) createDefaultNamespace(ctx context.Context, username string) error {
	span, _ := sg.FromGRPCContext(ctx, "ns create")
	defer span.Finish()
	fmt.Println(span)

	token, err := helper.ExtractToken(ctx)
	fmt.Println("{{BEFORE SENT TOKEN}}", token)
	if err != nil {
		span.AddLog(&sg.KV{"token error", err.Error()})
		return err
	}

	client := NewMeridianClient(s.meridian)
	_, err = client.Mutate(
		helper.AppendToken(
			sg.NewTracedGRPCContext(ctx, span),
			token,
		),
		&cPb.MutateReq{Mutate: &rPb.Task{
			UserId:    username,
			Namespace: "default",
			Extras: map[string]string{
				"namespace": "default",
			},
		}},
	)
	if err != nil {
		span.AddLog(&sg.KV{"meridian exists error", err.Error()})
		return err
	}
	return nil
}

func listOpt(req *cPb.ListReq, token string) *aPb.AuthOpt {
	return &aPb.AuthOpt{
		Data: map[string]string{
			"intent": "auth",
			"action": "list",
			"kind":   "roles",
			"token":  token,
		},
		Extras: map[string]*aPb.OptExtras{
			"user":      &aPb.OptExtras{Data: []string{req.Extras["user"]}},
			"namespace": &aPb.OptExtras{Data: []string{req.Extras["namespace"]}},
			"cmp":       &aPb.OptExtras{Data: []string{req.Extras["compare"]}},
			"labels":    &aPb.OptExtras{Data: []string{req.Extras["labels"]}},
		},
	}
}

func mutateOpt(req *cPb.MutateReq, token string) *aPb.AuthOpt {
	return &aPb.AuthOpt{
		Data: map[string]string{
			"intent":    "auth",
			"action":    "mutate",
			"kind":      "roles",
			"user":      req.Mutate.UserId,
			"token":     token,
			"namespace": req.Mutate.Namespace,
		},
		Extras: map[string]*aPb.OptExtras{
			"user":     &aPb.OptExtras{Data: []string{req.Mutate.Extras["user"]}},
			"resource": &aPb.OptExtras{Data: []string{req.Mutate.Extras["resources"]}},
			"verbs":    &aPb.OptExtras{Data: []string{req.Mutate.Extras["verbs"]}},
		},
	}
}
