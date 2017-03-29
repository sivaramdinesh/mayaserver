// This file handles jiva storage logic related to mayaserver's orchestration
// provider.
//
// NOTE:
//    jiva storage delegates the provisioning, placement & other operational
// aspects to an orchestration provider. Some of the orchestration providers
// can be Kubernetes, Nomad, etc.
package jiva

import (
	"fmt"

	"github.com/openebs/mayaserver/lib/api/v1"
	v1jiva "github.com/openebs/mayaserver/lib/api/v1/jiva"
	"github.com/openebs/mayaserver/lib/nethelper"
	"github.com/openebs/mayaserver/lib/volume"
	"k8s.io/apimachinery/pkg/api/resource"
)

type JivaInterface interface {

	// Name provides the name of the JivaInterface implementor
	Name() string

	// This is a builder method for NetworkOps. It will return false
	// if Network operations is not supported.
	NetworkOps() (NetworkOps, bool)

	// This is a builder method for StorageOps. It will return false
	// if Storage operations is not supported.
	StorageOps() (StorageOps, bool)
}

type NetworkOps interface {

	// NetworkProps does not fall under CRUD operations. This is applicable
	// to fetching properties from a config, or database etc.
	//
	// NOTE:
	//    This interface will have no control over Create, Update, Delete operations
	// of network properties
	NetworkProps(dc string) (map[v1.ContainerNetworkingLbl]string, error)
}

type StorageOps interface {

	// CRUD operations
	StorageInfo(*v1.PersistentVolumeClaim) (*v1.PersistentVolume, error)

	ProvisionStorage(*v1.PersistentVolumeClaim) (*v1.PersistentVolume, error)

	DeleteStorage(*v1.PersistentVolume) (*v1.PersistentVolume, error)

	// StorageProps does not fall under CRUD operations. This is applicable
	// to fetching properties from a config, or database etc.
	//
	// NOTE:
	//    This interface will have no control over Create, Update, Delete operations
	// of storage properties.
	//
	// NOTE:
	//    jiva requires these persistent storage properties to provision
	// its instances e.g. backing persistence location is required on which
	// a jiva replica can operate.
	StorageProps(dc string) (map[v1.ContainerStorageLbl]string, error)
}

// jivaUtil is the concrete implementation for
//
//  1. JivaInterface interface
//  2. NetworkOps interface
//  3. StorageOps interface
type jivaUtil struct {
	// Orthogonal concerns and their management w.r.t jiva storage
	// is done via aspect
	aspect volume.VolumePluginAspect
}

// newJivaUtil provides a orchestrator based infrastructure that
// supports jiva operations
func newJivaUtil(aspect volume.VolumePluginAspect) (JivaInterface, error) {
	if aspect == nil {
		return nil, fmt.Errorf("Nil volume plugin aspect was provided")
	}

	return &jivaUtil{
		aspect: aspect,
	}, nil
}

// This is a plain jiva utility implementation. Hence the name.
func (j *jivaUtil) Name() string {
	return "jivautil"
}

// StorageOps method provides an instance of StorageOps interface
//
// NOTE:
//  jivaUtil implements StorageOps interface. Hence it returns self.
func (j *jivaUtil) StorageOps() (StorageOps, bool) {
	return j, true
}

// NetworkOps method provides an instance of NetworkOps interface
//
// NOTE:
//  jivaUtil implements NetworkOps interface. Hence it returns self.
func (j *jivaUtil) NetworkOps() (NetworkOps, bool) {
	return j, true
}

// NetworkProps tries to fetch networking details from its orchestrator
//
// NOTE:
//  This is a concrete implementation of NetworkOps interface
func (j *jivaUtil) NetworkProps(dc string) (map[v1.ContainerNetworkingLbl]string, error) {

	orchestrator, err := j.aspect.GetOrchProvider()
	if err != nil {
		return nil, err
	}

	networkOrchestrator, ok := orchestrator.NetworkPlacements()

	if !ok {
		return nil, fmt.Errorf("Network operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return networkOrchestrator.NetworkPropsReq(dc)
}

// StorageProps tries to fetch persistent storage details from its orchestrator
//
// NOTE:
//  This is a concrete implementation of StorageOps interface
func (j *jivaUtil) StorageProps(dc string) (map[v1.ContainerStorageLbl]string, error) {

	orchestrator, err := j.aspect.GetOrchProvider()
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StoragePlacements()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.StoragePropsReq(dc)
}

// Info tries to fetch details of a jiva volume placed in an orchestrator
func (j *jivaUtil) StorageInfo(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {

	orchestrator, err := j.aspect.GetOrchProvider()
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StoragePlacements()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.StorageInfoReq(pvc)
}

// Provision tries to creates a jiva volume via an orchestrator
func (j *jivaUtil) ProvisionStorage(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolume, error) {

	orchestrator, err := j.aspect.GetOrchProvider()
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StoragePlacements()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	err = initLabels(pvc)
	if err != nil {
		return nil, err
	}

	err = verifySpecs(pvc)
	if err != nil {
		return nil, err
	}

	err = setJivaLblProps(pvc)
	if err != nil {
		return nil, err
	}

	err = setJivaSpecProps(pvc)
	if err != nil {
		return nil, err
	}

	err = j.setRegion(pvc)
	if err != nil {
		return nil, err
	}

	dc, err := j.setDC(pvc)
	if err != nil {
		return nil, err
	}

	err = j.setCN(dc, pvc)
	if err != nil {
		return nil, err
	}

	err = j.setCS(dc, pvc)
	if err != nil {
		return nil, err
	}

	return storageOrchestrator.StoragePlacementReq(pvc)
}

// initLabels is a utility function that will initialize the Labels
// of a PersistentVolumeClaim if not done so already.
func initLabels(pvc *v1.PersistentVolumeClaim) error {

	// return if already initialized
	if pvc.Labels != nil {
		return nil
	}

	// initialize with an empty list
	pvc.Labels = map[string]string{}

	return nil
}

// verifySpecs is a utility function that will verify the Spec
// of a PersistentVolumeClaim.
func verifySpecs(pvc *v1.PersistentVolumeClaim) error {

	if &pvc.Spec == nil || &pvc.Spec.Resources == nil || pvc.Spec.Resources.Requests == nil {
		return fmt.Errorf("Storage specs missing in pvc")
	}

	return nil
}

// setJivaLblProps function sets jiva specific properties with defaults
// if not done so already.
func setJivaLblProps(pvc *v1.PersistentVolumeClaim) error {

	if pvc.Labels == nil {
		return fmt.Errorf("Labels missing in pvc")
	}

	if pvc.Labels[string(v1jiva.JivaFrontEndImageLbl)] == "" {
		// TODO
		// Move to constants
		pvc.Labels[string(v1jiva.JivaFrontEndImageLbl)] = "openebs/jiva:latest"
	}

	return nil
}

// setJivaSpecProps function sets jiva specific properties with defaults
// if not done so already.
func setJivaSpecProps(pvc *v1.PersistentVolumeClaim) error {

	// Controller / Front End vol size
	feQuantity := pvc.Spec.Resources.Requests[v1jiva.JivaFrontEndVolSizeLbl]
	feQuantityPtr := &feQuantity

	if feQuantityPtr == nil || (feQuantityPtr != nil && feQuantityPtr.Sign() <= 0) {

		size, err := getStorageSize(pvc)
		if err != nil {
			return err
		}

		pvc.Spec.Resources.Requests[v1jiva.JivaFrontEndVolSizeLbl] = size
	}

	// Replica / Back End vol size
	beQuantity := pvc.Spec.Resources.Requests[v1jiva.JivaBackEndVolSizeLbl]
	beQuantityPtr := &beQuantity

	if beQuantityPtr == nil || (beQuantityPtr != nil && beQuantityPtr.Sign() <= 0) {

		size, err := getStorageSize(pvc)
		if err != nil {
			return err
		}

		pvc.Spec.Resources.Requests[v1jiva.JivaBackEndVolSizeLbl] = size
	}

	return nil
}

// getStorageSize gets the size of the storage if it was specified in
// persistent volume claim
func getStorageSize(pvc *v1.PersistentVolumeClaim) (resource.Quantity, error) {

	size := pvc.Spec.Resources.Requests["storage"]
	sizePtr := &size

	if sizePtr == nil {
		return size, fmt.Errorf("Storage size missing in pvc")
	}

	if sizePtr.Sign() <= 0 {
		return size, fmt.Errorf("Invalid storage size in pvc")
	}

	return size, nil
}

// setRegion sets the region property of a PersistentVolumeClaim
// if not done so already.
func (j *jivaUtil) setRegion(pvc *v1.PersistentVolumeClaim) error {

	if pvc.Labels == nil {
		return fmt.Errorf("Persistent volume claim's labels not initialized")
	}

	// return if region is already set
	if pvc.Labels[string(v1.RegionLbl)] != "" {
		return nil
	}

	if j.aspect == nil {
		return fmt.Errorf("Aspect missing while setting pvc region")
	}

	o, err := j.aspect.GetOrchProvider()
	if err != nil {
		return err
	}

	if o == nil {
		return fmt.Errorf("Orchestrator missing while setting pvc region")
	}

	// Set the pvc's region from jiva's orchestrator
	region := o.Region()
	if region == "" {
		return fmt.Errorf("Region could not be determined")
	}

	// set dc in pvc
	pvc.Labels[string(v1.RegionLbl)] = region

	return nil
}

// setDC sets the datacenter property of a PersistentVolumeClaim
// if not done so already.
func (j *jivaUtil) setDC(pvc *v1.PersistentVolumeClaim) (string, error) {

	if pvc.Labels == nil {
		return "", fmt.Errorf("Persistent volume claim's labels not initialized")
	}

	// return if dc is already set
	if pvc.Labels[string(v1.DatacenterLbl)] != "" {
		return pvc.Labels[string(v1.DatacenterLbl)], nil
	}

	// Set the pvc with dc from jiva's aspect
	dc, err := j.aspect.DefaultDatacenter()
	if err != nil {
		return "", err
	}

	if dc == "" {
		return "", fmt.Errorf("Datacenter could not be determined")
	}

	// set dc in pvc
	pvc.Labels[string(v1.DatacenterLbl)] = dc

	return dc, nil
}

// setCS sets the container storage properties in a PersistentVolumeClaim
// if not done so already.
func (j *jivaUtil) setCS(dc string, pvc *v1.PersistentVolumeClaim) error {

	if pvc.Labels == nil {
		return fmt.Errorf("Persistent volume claim's labels not initialized")
	}

	if dc == "" {
		return fmt.Errorf("Datacenter not provided")
	}

	// Fetch the networking options that are orchestrator & datacenter specific
	cs, err := j.StorageProps(dc)
	if err != nil {
		return err
	}

	// Set the persistent storage options if not already set
	//
	// NOTE:
	//    User provided persistent storage options score over
	// orchestrator & dc specific configurations
	for k, _ := range cs {
		if pvc.Labels[string(k)] == "" {
			pvc.Labels[string(k)] = cs[k]
		}
	}

	return nil
}

// setCN sets the container networking properties in a PersistentVolumeClaim
// if not done so already.
func (j *jivaUtil) setCN(dc string, pvc *v1.PersistentVolumeClaim) error {

	if pvc.Labels == nil {
		return fmt.Errorf("Persistent volume claim's labels not initialized")
	}

	if dc == "" {
		return fmt.Errorf("Datacenter not provided")
	}

	// Fetch the networking options that are orchestrator & datacenter specific
	cn, err := j.NetworkProps(dc)
	if err != nil {
		return err
	}

	// Set the networking options if not already set
	//
	// NOTE:
	//    User provided networking options score over
	// orchestrator & dc specific configurations
	for k, _ := range cn {
		if pvc.Labels[string(k)] == "" {
			pvc.Labels[string(k)] = cn[k]
		}
	}

	networkCIDR := pvc.Labels[string(v1.CNNetworkCIDRAddrLbl)]
	if networkCIDR == "" {
		return fmt.Errorf("Network CIDR could not be determined")
	}

	subnet, err := nethelper.CIDRSubnet(networkCIDR)
	if err != nil {
		return err
	}

	pvc.Labels[string(v1.CNSubnetLbl)] = subnet

	// TODO
	// Below is a tightly coupled design which makes sense of only
	// one jiva controller & one jiva replica.
	if pvc.Labels[string(v1jiva.JivaFrontEndIPLbl)] == "" && pvc.Labels[string(v1jiva.JivaBackEndAllIPsLbl)] == "" {
		// Get three available IPs
		ips, err := nethelper.GetAvailableIPs(networkCIDR, 3)
		if err != nil {
			return err
		}

		pvc.Labels[string(v1jiva.JivaFrontEndIPLbl)] = ips[0]
		pvc.Labels[string(v1jiva.JivaBackEndAllIPsLbl)] = ips[1] + "," + ips[2]

		return nil
	}

	if pvc.Labels[string(v1jiva.JivaFrontEndIPLbl)] == "" {
		// Get one available IP
		ips, err := nethelper.GetAvailableIPs(networkCIDR, 1)
		if err != nil {
			return err
		}

		pvc.Labels[string(v1jiva.JivaFrontEndIPLbl)] = ips[0]

		return nil
	}

	if pvc.Labels[string(v1jiva.JivaBackEndAllIPsLbl)] == "" {
		// Get two available IPs
		ips, err := nethelper.GetAvailableIPs(networkCIDR, 2)
		if err != nil {
			return err
		}

		pvc.Labels[string(v1jiva.JivaBackEndAllIPsLbl)] = ips[0] + "," + ips[1]

		return nil
	}

	return nil
}

// Delete tries to delete the jiva volume via an orchestrator
func (j *jivaUtil) DeleteStorage(pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	orchestrator, err := j.aspect.GetOrchProvider()
	if err != nil {
		return nil, err
	}

	storageOrchestrator, ok := orchestrator.StoragePlacements()

	if !ok {
		return nil, fmt.Errorf("Storage operations not supported by orchestrator '%s'", orchestrator.Name())
	}

	return storageOrchestrator.StorageRemovalReq(pv)
}
