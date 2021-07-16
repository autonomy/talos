// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package secrets

import (
	"context"
	stdlibx509 "crypto/x509"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/talos-systems/crypto/x509"
	"go.uber.org/zap"

	"github.com/talos-systems/talos/pkg/grpc/gen"
	"github.com/talos-systems/talos/pkg/machinery/config/types/v1alpha1/machine"
	"github.com/talos-systems/talos/pkg/machinery/role"
	"github.com/talos-systems/talos/pkg/resources/config"
	"github.com/talos-systems/talos/pkg/resources/k8s"
	"github.com/talos-systems/talos/pkg/resources/network"
	"github.com/talos-systems/talos/pkg/resources/secrets"
	timeresource "github.com/talos-systems/talos/pkg/resources/time"
	"github.com/talos-systems/talos/pkg/resources/v1alpha1"
)

// APIController manages secrets.API based on configuration to provide apid certificate.
type APIController struct{}

// Name implements controller.Controller interface.
func (ctrl *APIController) Name() string {
	return "secrets.APIController"
}

// Inputs implements controller.Controller interface.
func (ctrl *APIController) Inputs() []controller.Input {
	// initial set of inputs: wait for machine type to be known and network to be partially configured
	return []controller.Input{
		{
			Namespace: network.NamespaceName,
			Type:      network.StatusType,
			ID:        pointer.ToString(network.StatusID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: config.NamespaceName,
			Type:      config.MachineTypeType,
			ID:        pointer.ToString(config.MachineTypeID),
			Kind:      controller.InputWeak,
		},
	}
}

// Outputs implements controller.Controller interface.
func (ctrl *APIController) Outputs() []controller.Output {
	return []controller.Output{
		{
			Type: secrets.APIType,
			Kind: controller.OutputExclusive,
		},
	}
}

// Run implements controller.Controller interface.
//
//nolint:gocyclo
func (ctrl *APIController) Run(ctx context.Context, r controller.Runtime, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		}

		machineTypeRes, err := r.Get(ctx, resource.NewMetadata(config.NamespaceName, config.MachineTypeType, config.MachineTypeID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return fmt.Errorf("error getting machine type: %w", err)
		}

		machineType := machineTypeRes.(*config.MachineType).MachineType()

		networkResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.StatusType, network.StatusID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		networkStatus := networkResource.(*network.Status).TypedSpec()

		if !(networkStatus.AddressReady && networkStatus.HostnameReady) {
			continue
		}

		// machine type is known and network is ready, we can now proceed to one or another reconcile loop
		switch machineType {
		case machine.TypeInit, machine.TypeControlPlane:
			if err = ctrl.reconcile(ctx, r, logger, true); err != nil {
				return err
			}
		case machine.TypeWorker:
			if err = ctrl.reconcile(ctx, r, logger, false); err != nil {
				return err
			}
		case machine.TypeUnknown:
			// machine configuration is not loaded yet, do nothing
		default:
			panic(fmt.Sprintf("unexpected machine type %v", machineType))
		}

		if err = ctrl.teardownAll(ctx, r); err != nil {
			return err
		}
	}
}

//nolint:gocyclo,cyclop
func (ctrl *APIController) reconcile(ctx context.Context, r controller.Runtime, logger *zap.Logger, isControlplane bool) error {
	inputs := []controller.Input{
		{
			Namespace: secrets.NamespaceName,
			Type:      secrets.RootType,
			ID:        pointer.ToString(secrets.RootOSID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: network.NamespaceName,
			Type:      network.HostnameStatusType,
			ID:        pointer.ToString(network.HostnameID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: network.NamespaceName,
			Type:      network.NodeAddressType,
			ID:        pointer.ToString(network.NodeAddressAccumulativeID),
			Kind:      controller.InputWeak,
		},
		{
			Namespace: config.NamespaceName,
			Type:      config.MachineTypeType,
			ID:        pointer.ToString(config.MachineTypeID),
			Kind:      controller.InputWeak,
		},
		// time status isn't fetched, but the fact that it is in dependencies means
		// that certs will be regenerated on time sync/jump (as reconcile will be triggered)
		{
			Namespace: v1alpha1.NamespaceName,
			Type:      timeresource.StatusType,
			ID:        pointer.ToString(timeresource.StatusID),
			Kind:      controller.InputWeak,
		},
	}

	if !isControlplane {
		// worker nodes depend on endpoint list
		inputs = append(inputs, controller.Input{
			Namespace: k8s.ControlPlaneNamespaceName,
			Type:      k8s.EndpointType,
			ID:        pointer.ToString(k8s.ControlPlaneEndpointsID),
			Kind:      controller.InputWeak,
		})
	}

	if err := r.UpdateInputs(inputs); err != nil {
		return fmt.Errorf("error updating inputs: %w", err)
	}

	r.QueueReconcile()

	refreshTicker := time.NewTicker(x509.DefaultCertificateValidityDuration / 2)
	defer refreshTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		case <-refreshTicker.C:
		}

		machineTypeRes, err := r.Get(ctx, resource.NewMetadata(config.NamespaceName, config.MachineTypeType, config.MachineTypeID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return fmt.Errorf("error getting machine type: %w", err)
		}

		machineType := machineTypeRes.(*config.MachineType).MachineType()

		switch machineType {
		case machine.TypeInit, machine.TypeControlPlane:
			if !isControlplane {
				return fmt.Errorf("machine type changed")
			}
		case machine.TypeWorker:
			if isControlplane {
				return fmt.Errorf("machine type changed")
			}
		case machine.TypeUnknown:
			return fmt.Errorf("machine type changed")
		default:
			panic(fmt.Sprintf("unexpected machine type %v", machineType))
		}

		rootResource, err := r.Get(ctx, resource.NewMetadata(secrets.NamespaceName, secrets.RootType, secrets.RootOSID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				if err = ctrl.teardownAll(ctx, r); err != nil {
					return fmt.Errorf("error destroying resources: %w", err)
				}

				continue
			}

			return fmt.Errorf("error getting etcd root secrets: %w", err)
		}

		rootSpec := rootResource.(*secrets.Root).OSSpec()

		hostnameResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.HostnameStatusType, network.HostnameID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		hostnameStatus := hostnameResource.(*network.HostnameStatus).TypedSpec()

		addressesResource, err := r.Get(ctx, resource.NewMetadata(network.NamespaceName, network.NodeAddressType, network.NodeAddressAccumulativeID, resource.VersionUndefined))
		if err != nil {
			if state.IsNotFoundError(err) {
				continue
			}

			return err
		}

		nodeAddresses := addressesResource.(*network.NodeAddress).TypedSpec()

		var endpointsStr []string

		if !isControlplane {
			endpointResource, err := r.Get(ctx, resource.NewMetadata(k8s.ControlPlaneNamespaceName, k8s.EndpointType, k8s.ControlPlaneEndpointsID, resource.VersionUndefined))
			if err != nil {
				if state.IsNotFoundError(err) {
					continue
				}

				return fmt.Errorf("error getting endpoints resource: %w", err)
			}

			endpoints := endpointResource.(*k8s.Endpoint).TypedSpec()

			if len(endpoints.Addresses) == 0 {
				continue
			}

			endpointsStr = make([]string, 0, len(endpoints.Addresses))

			for _, ip := range endpoints.Addresses {
				endpointsStr = append(endpointsStr, ip.String())
			}
		}

		var altNames AltNames

		for _, ip := range append(rootSpec.CertSANIPs, nodeAddresses.Addresses...) {
			altNames.AppendIPs(ip.IPAddr().IP)
		}

		altNames.AppendDNSNames(rootSpec.CertSANDNSNames...)
		altNames.AppendDNSNames(hostnameStatus.Hostname, hostnameStatus.FQDN())

		if isControlplane {
			if err := ctrl.generateControlPlane(ctx, r, logger, rootSpec, altNames, hostnameStatus.FQDN()); err != nil {
				return err
			}
		} else {
			if err := ctrl.generateJoin(ctx, r, logger, rootSpec, endpointsStr, altNames, hostnameStatus.FQDN()); err != nil {
				return err
			}
		}
	}
}

func (ctrl *APIController) generateControlPlane(ctx context.Context, r controller.Runtime, logger *zap.Logger, rootSpec *secrets.RootOSSpec, altNames AltNames, fqdn string) error {
	ca, err := x509.NewCertificateAuthorityFromCertificateAndKey(rootSpec.CA)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	serverCert, err := x509.NewKeyPair(ca,
		x509.IPAddresses(altNames.IPs),
		x509.DNSNames(altNames.DNSNames),
		x509.CommonName(fqdn),
		x509.NotAfter(time.Now().Add(x509.DefaultCertificateValidityDuration)),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageServerAuth,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate API server cert: %w", err)
	}

	clientCert, err := x509.NewKeyPair(ca,
		x509.CommonName(fqdn),
		x509.Organization(string(role.Impersonator)),
		x509.NotAfter(time.Now().Add(x509.DefaultCertificateValidityDuration)),
		x509.KeyUsage(stdlibx509.KeyUsageDigitalSignature|stdlibx509.KeyUsageKeyEncipherment),
		x509.ExtKeyUsage([]stdlibx509.ExtKeyUsage{
			stdlibx509.ExtKeyUsageClientAuth,
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to generate API client cert: %w", err)
	}

	if err := r.Modify(ctx, secrets.NewAPI(),
		func(r resource.Resource) error {
			apiSecrets := r.(*secrets.API).TypedSpec()

			apiSecrets.CA = &x509.PEMEncodedCertificateAndKey{
				Crt: rootSpec.CA.Crt,
			}
			apiSecrets.Server = x509.NewCertificateAndKeyFromKeyPair(serverCert)
			apiSecrets.Client = x509.NewCertificateAndKeyFromKeyPair(clientCert)

			return nil
		}); err != nil {
		return fmt.Errorf("error modifying resource: %w", err)
	}

	clientFingerprint, _ := x509.SPKIFingerprintFromDER(clientCert.Certificate.Certificate[0]) //nolint:errcheck
	serverFingerprint, _ := x509.SPKIFingerprintFromDER(serverCert.Certificate.Certificate[0]) //nolint:errcheck

	logger.Debug("generated new certificates",
		zap.Stringer("client", clientFingerprint),
		zap.Stringer("server", serverFingerprint),
	)

	return nil
}

func (ctrl *APIController) generateJoin(ctx context.Context, r controller.Runtime, logger *zap.Logger,
	rootSpec *secrets.RootOSSpec, endpointsStr []string, altNames AltNames, fqdn string) error {
	remoteGen, err := gen.NewRemoteGenerator(rootSpec.Token, endpointsStr)
	if err != nil {
		return fmt.Errorf("failed creating trustd client: %w", err)
	}

	defer remoteGen.Close() //nolint:errcheck

	serverCSR, serverCert, err := x509.NewEd25519CSRAndIdentity(
		x509.IPAddresses(altNames.IPs),
		x509.DNSNames(altNames.DNSNames),
		x509.CommonName(fqdn),
	)
	if err != nil {
		return fmt.Errorf("failed to generate API server CSR: %w", err)
	}

	var ca []byte

	ca, serverCert.Crt, err = remoteGen.IdentityContext(ctx, serverCSR)
	if err != nil {
		return fmt.Errorf("failed to sign API server CSR: %w", err)
	}

	clientCSR, clientCert, err := x509.NewEd25519CSRAndIdentity(
		x509.CommonName(fqdn),
		x509.Organization(string(role.Impersonator)),
	)
	if err != nil {
		return fmt.Errorf("failed to generate API client CSR: %w", err)
	}

	// TODO: add keyusage: trustd should accept key usage as additional params
	_, clientCert.Crt, err = remoteGen.IdentityContext(ctx, clientCSR)
	if err != nil {
		return fmt.Errorf("failed to sign API client CSR: %w", err)
	}

	if err := r.Modify(ctx, secrets.NewAPI(),
		func(r resource.Resource) error {
			apiSecrets := r.(*secrets.API).TypedSpec()

			apiSecrets.CA = &x509.PEMEncodedCertificateAndKey{
				Crt: ca,
			}
			apiSecrets.Server = serverCert
			apiSecrets.Client = clientCert

			return nil
		}); err != nil {
		return fmt.Errorf("error modifying resource: %w", err)
	}

	clientFingerprint, _ := x509.SPKIFingerprintFromPEM(clientCert.Crt) //nolint:errcheck
	serverFingerprint, _ := x509.SPKIFingerprintFromPEM(serverCert.Crt) //nolint:errcheck

	logger.Debug("generated new certificates",
		zap.Stringer("client", clientFingerprint),
		zap.Stringer("server", serverFingerprint),
	)

	return nil
}

func (ctrl *APIController) teardownAll(ctx context.Context, r controller.Runtime) error {
	list, err := r.List(ctx, resource.NewMetadata(secrets.NamespaceName, secrets.APIType, "", resource.VersionUndefined))
	if err != nil {
		return err
	}

	// TODO: change this to proper teardown sequence

	for _, res := range list.Items {
		if err = r.Destroy(ctx, res.Metadata()); err != nil {
			return err
		}
	}

	return nil
}
