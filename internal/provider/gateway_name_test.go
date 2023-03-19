// Package provider is the terraform provider
package provider

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/substrate-client"
	client "github.com/threefoldtech/terraform-provider-grid/internal/node"
	mock "github.com/threefoldtech/terraform-provider-grid/internal/provider/mocks"
	"github.com/threefoldtech/terraform-provider-grid/pkg/subi"
	"github.com/threefoldtech/terraform-provider-grid/pkg/workloads"
	"github.com/threefoldtech/zos/pkg/gridtypes"
	"github.com/threefoldtech/zos/pkg/gridtypes/zos"
)

func TestNameValidateNodeNotReachable(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sub := mock.NewMockSubstrateExt(ctrl)
	cl := mock.NewRMBMockClient(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	cl.
		EXPECT().
		Call(
			gomock.Any(),
			uint32(10),
			"zos.system.version",
			nil,
			gomock.Any(),
		).
		Return(errors.New("couldn't reach node"))
	pool.
		EXPECT().
		GetNodeClient(
			gomock.Any(),
			uint32(11),
		).
		Return(client.NewNodeClient(10, cl, 10*time.Second), nil)

	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
		},
		ncPool: pool,
		Node:   11,
	}
	err = gw.Validate(context.TODO(), sub)
	assert.Error(t, err)
}

func TestNameValidateNodeReachable(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	sub := mock.NewMockSubstrateExt(ctrl)
	cl := mock.NewRMBMockClient(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	cl.
		EXPECT().
		Call(
			gomock.Any(),
			uint32(10),
			"zos.system.version",
			nil,
			gomock.Any(),
		).
		Return(nil)
	pool.
		EXPECT().
		GetNodeClient(
			gomock.Any(),
			uint32(11),
		).
		Return(client.NewNodeClient(10, cl, 10*time.Second), nil)

	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
		},
		ncPool: pool,
		Node:   11,
	}
	err = gw.Validate(context.TODO(), sub)
	assert.NoError(t, err)
}

func TestNameGenerateDeployment(t *testing.T) {
	g := workloads.GatewayNameProxy{
		Name:           "name",
		TLSPassthrough: false,
		Backends:       []zos.Backend{"a", "b"},
		FQDN:           "name.com",
	}
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			twinID: 11,
		},
		Node: 10,
		Gw:   g,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, dls, map[uint32]gridtypes.Deployment{
		10: {
			Version: 0,
			TwinID:  11,
			Workloads: []gridtypes.Workload{
				{
					Version: 0,
					Type:    zos.GatewayNameProxyType,
					Name:    gridtypes.Name(g.Name),
					Data: gridtypes.MustMarshal(zos.GatewayNameProxy{
						GatewayBase: zos.GatewayBase{
							TLSPassthrough: g.TLSPassthrough,
							Backends:       g.Backends,
						},
						Name: g.Name,
					}),
				},
			},
			SignatureRequirement: gridtypes.SignatureRequirement{
				WeightRequired: 1,
				Requests: []gridtypes.SignatureRequest{
					{
						TwinID: 11,
						Weight: 1,
					},
				},
			},
		},
	})
}

func TestNameDeploy(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	cl := mock.NewRMBMockClient(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)

	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		ncPool:   pool,
		deployer: deployer,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		nil,
		dls,
	).Return(map[uint32]uint64{10: 100}, nil)
	sub.EXPECT().
		CreateNameContract(identity, "name").
		Return(uint64(100), nil)
	pool.EXPECT().
		GetNodeClient(sub, uint32(10)).
		Return(client.NewNodeClient(12, cl, 10*time.Second), nil)
	cl.EXPECT().Call(
		gomock.Any(),
		uint32(12),
		"zos.system.version",
		gomock.Any(),
		gomock.Any(),
	).Return(nil)
	err = gw.Deploy(context.Background(), sub)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
}

func TestNameUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	cl := mock.NewRMBMockClient(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		deployer:         deployer,
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
		ncPool:           pool,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		map[uint32]uint64{10: 100},
		dls,
	).Return(map[uint32]uint64{10: 100}, nil)
	sub.EXPECT().
		InvalidateNameContract(gomock.Any(), identity, uint64(200), gw.Gw.Name).
		Return(uint64(200), nil)

	pool.EXPECT().
		GetNodeClient(sub, uint32(10)).
		Return(client.NewNodeClient(12, cl, 10*time.Second), nil)
	cl.EXPECT().Call(
		gomock.Any(),
		uint32(12),
		"zos.system.version",
		gomock.Any(),
		gomock.Any(),
	).Return(nil)
	err = gw.Deploy(context.Background(), sub)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{uint32(10): uint64(100)})
}

func TestNameUpdateFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	cl := mock.NewRMBMockClient(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		deployer:         deployer,
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
		ncPool:           pool,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		map[uint32]uint64{10: 100},
		dls,
	).Return(map[uint32]uint64{10: 100}, errors.New("error"))
	sub.EXPECT().
		InvalidateNameContract(gomock.Any(), identity, uint64(200), gw.Gw.Name).
		Return(uint64(200), nil)
	pool.EXPECT().
		GetNodeClient(sub, uint32(10)).
		Return(client.NewNodeClient(12, cl, 10*time.Second), nil)
	cl.EXPECT().Call(
		gomock.Any(),
		uint32(12),
		"zos.system.version",
		gomock.Any(),
		gomock.Any(),
	).Return(nil)

	err = gw.Deploy(context.Background(), sub)
	assert.Error(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{uint32(10): uint64(100)})
	assert.Equal(t, gw.NameContractID, uint64(200))
}

func TestNameCancel(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		deployer:         deployer,
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
	}
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		map[uint32]uint64{10: 100},
		map[uint32]gridtypes.Deployment{},
	).Return(map[uint32]uint64{}, nil)
	sub.EXPECT().
		EnsureContractCanceled(identity, uint64(200)).
		Return(nil)

	err = gw.Cancel(context.Background(), sub)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{})
	assert.Equal(t, gw.NameContractID, uint64(0))
}

func TestNameCancelDeploymentsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		deployer:         deployer,
		NodeDeploymentID: map[uint32]uint64{10: 100},
	}
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		map[uint32]uint64{10: 100},
		map[uint32]gridtypes.Deployment{},
	).Return(map[uint32]uint64{10: 100}, errors.New("error"))
	err = gw.Cancel(context.Background(), sub)
	assert.Error(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
}

func TestNameCancelContractsFailed(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		deployer:         deployer,
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
	}
	deployer.EXPECT().Deploy(
		gomock.Any(),
		sub,
		map[uint32]uint64{10: 100},
		map[uint32]gridtypes.Deployment{},
	).Return(map[uint32]uint64{}, nil)
	sub.EXPECT().
		EnsureContractCanceled(identity, uint64(200)).
		Return(errors.New("error"))

	err = gw.Cancel(context.Background(), sub)
	assert.Error(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{})
	assert.Equal(t, gw.NameContractID, uint64(200))
}

func TestNameSyncContracts(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ID: "123",
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
	}
	sub.EXPECT().DeleteInvalidContracts(
		gw.NodeDeploymentID,
	).Return(nil)
	sub.EXPECT().IsValidContract(
		gw.NameContractID,
	).Return(true, nil)

	err = gw.syncContracts(context.Background(), sub)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
	assert.Equal(t, gw.ID, "123")
}

func TestNameSyncDeletedContracts(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ID: "123",
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
	}
	sub.EXPECT().DeleteInvalidContracts(
		gw.NodeDeploymentID,
	).DoAndReturn(func(contracts map[uint32]uint64) error {
		delete(contracts, 10)
		return nil
	})
	sub.EXPECT().IsValidContract(
		gw.NameContractID,
	).Return(false, nil)
	err = gw.syncContracts(context.Background(), sub)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{})
	assert.Equal(t, gw.NameContractID, uint64(0))
	assert.Equal(t, gw.ID, "")
}

func TestNameSyncContractsFailure(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ID: "123",
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
	}
	sub.EXPECT().DeleteInvalidContracts(
		gw.NodeDeploymentID,
	).Return(errors.New("123"))

	err = gw.syncContracts(context.Background(), sub)
	assert.Error(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
	assert.Equal(t, gw.NameContractID, uint64(200))
	assert.Equal(t, gw.ID, "123")
}

func TestNameSync(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ID: "123",
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		NodeDeploymentID: map[uint32]uint64{10: 100},
		NameContractID:   200,
		deployer:         deployer,
		ncPool:           pool,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	dl := dls[10]
	dl.Workloads[0].Result.State = gridtypes.StateOk
	dl.Workloads[0].Result.Data, err = json.Marshal(zos.GatewayProxyResult{FQDN: "name.com"})
	assert.NoError(t, err)
	sub.EXPECT().DeleteInvalidContracts(
		gw.NodeDeploymentID,
	).Return(nil)
	sub.EXPECT().IsValidContract(
		gw.NameContractID,
	).Return(true, nil)

	deployer.EXPECT().
		GetDeployments(gomock.Any(), sub, map[uint32]uint64{10: 100}).
		DoAndReturn(func(ctx context.Context, _ subi.SubstrateExt, _ map[uint32]uint64) (map[uint32]gridtypes.Deployment, error) {
			return map[uint32]gridtypes.Deployment{10: dl}, nil
		})
	gw.Gw.FQDN = "123"
	err = gw.Sync(context.Background(), sub, gw.ThreefoldPluginClient)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
	assert.Equal(t, gw.NameContractID, uint64(200))
	assert.Equal(t, gw.ID, "123")
	assert.Equal(t, gw.Gw.FQDN, "name.com")
}

func TestNameSyncDeletedWorkload(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	identity, err := substrate.NewIdentityFromEd25519Phrase(Words)
	assert.NoError(t, err)
	deployer := mock.NewMockDeployer(ctrl)
	pool := mock.NewMockNodeClientGetter(ctrl)
	sub := mock.NewMockSubstrateExt(ctrl)
	gw := GatewayNameDeployer{
		ID: "123",
		ThreefoldPluginClient: &threefoldPluginClient{
			identity: identity,
			twinID:   11,
		},
		Node: 10,
		Gw: workloads.GatewayNameProxy{
			Name:           "name",
			TLSPassthrough: false,
			Backends:       []zos.Backend{"https://1.1.1.1", "http://2.2.2.2"},
			FQDN:           "name.com",
		},
		NodeDeploymentID: map[uint32]uint64{10: 100},
		deployer:         deployer,
		ncPool:           pool,
	}
	dls, err := gw.GenerateVersionlessDeployments(context.Background())
	assert.NoError(t, err)
	dl := dls[10]
	// state is deleted

	sub.EXPECT().DeleteInvalidContracts(
		gw.NodeDeploymentID,
	).Return(nil)
	sub.EXPECT().IsValidContract(
		gw.NameContractID,
	).Return(true, nil)

	deployer.EXPECT().
		GetDeployments(gomock.Any(), sub, map[uint32]uint64{10: 100}).
		DoAndReturn(func(ctx context.Context, _ subi.SubstrateExt, _ map[uint32]uint64) (map[uint32]gridtypes.Deployment, error) {
			return map[uint32]gridtypes.Deployment{10: dl}, nil
		})
	gw.Gw.FQDN = "123"
	err = gw.Sync(context.Background(), sub, gw.ThreefoldPluginClient)
	assert.NoError(t, err)
	assert.Equal(t, gw.NodeDeploymentID, map[uint32]uint64{10: 100})
	assert.Equal(t, gw.ID, "123")
	assert.Equal(t, gw.Gw, workloads.GatewayNameProxy{})
}
