package jiva

import (
	"github.com/openebs/mayaserver/lib/api/v1"
	v1nomad "github.com/openebs/mayaserver/lib/api/v1/nomad"
)

// These label can be used in the persistent volume claims to specify
// exact requirements.
type JivaLbl string

const (
	JivaFrontEndImageLbl JivaLbl = "fe.jiva.volume.openebs.io/image-version"
	JivaBackEndImageLbl  JivaLbl = "be.jiva.volume.openebs.io/image-version"

	JivaFrontEndIPLbl JivaLbl = "fe.jiva.volume.openebs.io/ip"
	JivaBackEndIPLbl  JivaLbl = "be.jiva.volume.openebs.io/ip"

	JivaFrontEndAllIPsLbl JivaLbl = "fe.jiva.volume.openebs.io/all-ips"
	JivaBackEndAllIPsLbl  JivaLbl = "be.jiva.volume.openebs.io/all-ips"

	JivaFrontEndCountLbl JivaLbl = "fe.jiva.volume.openebs.io/count"
	JivaBackEndCountLbl  JivaLbl = "be.jiva.volume.openebs.io/count"

	JivaBackEndVolStor JivaLbl = "be.jiva.volume.openebs.io/vol-stor"
)

const (
	JivaFrontEndVolSizeLbl v1.ResourceName = "fe.jiva.volume.openebs.io/vol-size"
	JivaBackEndVolSizeLbl  v1.ResourceName = "be.jiva.volume.openebs.io/vol-size"
)

const (
	// This naming is a result of considering Jiva volume plugin's default
	// orchestrator which is Nomad & this default orchestrator's default region
	// which is global.
	DefaultJivaVolumePluginName string = v1.VolumePluginNamePrefix + "jiva-nomad-" + v1nomad.DefaultNomadRegionName

	// This just points to Nomad orchestrator's default dc.
	DefaultJivaDataCenter string = v1nomad.DefaultNomadDCName

	// This naming signifies a prefix that can be used with variants of
	// jiva volume plugin.
	//
	// NOTE:
	// Some sample variants of jiva volume plugin:
	//
	//  Jiva + K8S + us-east-1
	//  Jiva + Nomad + global
	//  Jiva + Nomad + in-bang
	JivaVolumePluginNamePrefix string = v1.VolumePluginNamePrefix + "jiva-"

	JivaIscsiTargetPortalPort string = "3260"

	// Jiva iSCSI Qualified Name format
	JivaIqnFormatPrefix string = "iqn.2016-09.com.openebs.jiva"
)
