package provider

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/threefoldtech/zos/client"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/gridtypes/zos"
	"github.com/threefoldtech/zos/pkg/substrate"
)

const (
	Version = 0
	// Twin      = 14
	NodeID = 2
	// Seed      = "d161de46d136d96085906b9f3d40d08b3649c80a3e4d77f0b14d3dc6889e9dcb"
	// Substrate = "wss://explorer.devnet.grid.tf/ws"
	// rmb_url   = "tcp://127.0.0.1:6379"
)

func resourceDisk() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Sample resource in the Terraform provider scaffolding.",

		CreateContext: resourceDiskCreate,
		ReadContext:   resourceScaffoldingRead,
		UpdateContext: resourceScaffoldingUpdate,
		DeleteContext: resourceScaffoldingDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Disk Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"version": {
				Description: "Version",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"description": {
				Description: "Description field",
				Type:        schema.TypeString,
				Required:    true,
			},
			"size": {
				Description: "Disk size in Gigabytes",
				Type:        schema.TypeInt,
				Required:    true,
			},
		},
	}
}

// func deploy(deployment []gridtypes.Workload, apiClient apiClient){

// }
func resourceDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)
	// seed, err := hex.DecodeString(Seed)
	// if err != nil {
	// 	panic(err)
	// }
	// userSK := ed25519.NewKeyFromSeed(seed)
	// cl, err := rmb.NewClient(rmb_url)
	// if err != nil {
	// 	panic(err)
	// }
	apiClient := meta.(*apiClient)
	userSK := ed25519.NewKeyFromSeed(apiClient.seed)
	cl := apiClient.client

	var diags diag.Diagnostics

	workload := gridtypes.Workload{
		Name:        gridtypes.Name(d.Get("name").(string)),
		Version:     Version,
		Type:        zos.ZMountType,
		Description: d.Get("description").(string),
		Data: gridtypes.MustMarshal(zos.ZMount{
			Size: gridtypes.Unit(d.Get("size").(int)) * gridtypes.Gigabyte,
		}),
	}

	dl := gridtypes.Deployment{
		Version: Version,
		TwinID:  uint32(apiClient.twin_id), //LocalTwin,
		// this contract id must match the one on substrate
		Workloads: []gridtypes.Workload{
			workload,
		},
		SignatureRequirement: gridtypes.SignatureRequirement{
			WeightRequired: 1,
			Requests: []gridtypes.SignatureRequest{
				{
					TwinID: apiClient.twin_id,
					Weight: 1,
				},
			},
		},
	}

	if err := dl.Valid(); err != nil {
		panic("invalid: " + err.Error())
	}
	//return
	if err := dl.Sign(apiClient.twin_id, userSK); err != nil {
		panic(err)
	}

	hash, err := dl.ChallengeHash()
	if err != nil {
		panic("failed to create hash")
	}

	hashHex := hex.EncodeToString(hash)
	fmt.Printf("hash: %s\n", hashHex)
	// create contract
	sub, err := substrate.NewSubstrate(apiClient.substrate_url)
	if err != nil {
		panic(err)
	}
	nodeInfo, err := sub.GetNode(NodeID)
	if err != nil {
		panic(err)
	}

	node := client.NewNodeClient(uint32(nodeInfo.TwinID), cl)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	total, used, err := node.Counters(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Total: %+v\nUsed: %+v\n", total, used)

	contractID, err := sub.CreateContract(userSK, NodeID, nil, hashHex, 1)
	if err != nil {
		panic(err)
	}
	dl.ContractID = contractID // from substrate

	err = node.DeploymentDeploy(ctx, dl)
	if err != nil {
		panic(err)
	}

	got, err := node.DeploymentGet(ctx, dl.ContractID)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(got)
	d.SetId(strconv.FormatUint(contractID, 10))
	return diags
}

func resourceDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	apiClient := meta.(*apiClient)
	userSK := ed25519.NewKeyFromSeed(apiClient.seed)
	cl := apiClient.client

	var diags diag.Diagnostics

	workload := gridtypes.Workload{
		Name:        gridtypes.Name(d.Get("name").(string)),
		Version:     Version,
		Type:        zos.ZMountType,
		Description: d.Get("description").(string),
		Data: gridtypes.MustMarshal(zos.ZMount{
			Size: gridtypes.Unit(d.Get("size").(int)) * gridtypes.Gigabyte,
		}),
	}
	dl := gridtypes.Deployment{
		Version: Version,
		TwinID:  apiClient.twin_id, //LocalTwin,
		// this contract id must match the one on substrate
		Workloads: []gridtypes.Workload{
			workload,
		},
		SignatureRequirement: gridtypes.SignatureRequirement{
			WeightRequired: 1,
			Requests: []gridtypes.SignatureRequest{
				{
					TwinID: apiClient.twin_id,
					Weight: 1,
				},
			},
		},
	}

	if err := dl.Valid(); err != nil {
		panic("invalid: " + err.Error())
	}
	//return
	if err := dl.Sign(apiClient.twin_id, userSK); err != nil {
		panic(err)
	}

	hash, err := dl.ChallengeHash()
	if err != nil {
		panic("failed to create hash")
	}

	hashHex := hex.EncodeToString(hash)
	fmt.Printf("hash: %s\n", hashHex)
	// create contract
	sub, err := substrate.NewSubstrate(apiClient.substrate_url)
	if err != nil {
		panic(err)
	}
	nodeInfo, err := sub.GetNode(NodeID)
	if err != nil {
		panic(err)
	}

	node := client.NewNodeClient(uint32(nodeInfo.TwinID), cl)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	total, used, err := node.Counters(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Total: %+v\nUsed: %+v\n", total, used)

	contractID, err := sub.CreateContract(userSK, NodeID, nil, hashHex, 1)
	if err != nil {
		panic(err)
	}
	dl.ContractID = contractID // from substrate

	err = node.DeploymentDeploy(ctx, dl)
	if err != nil {
		panic(err)
	}

	got, err := node.DeploymentGet(ctx, dl.ContractID)
	if err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(got)
	d.SetId(strconv.FormatUint(contractID, 10))
	return diags
}

func resourceDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}

func resourceDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	// client := meta.(*apiClient)

	return diag.Errorf("not implemented")
}