package kubetest

import (
	"sync"

	osapps_v1 "github.com/openshift/api/apps/v1"
	"github.com/stretchr/testify/mock"
	"gopkg.in/square/go-jose.v2/jwt"
	istio_fake "istio.io/client-go/pkg/clientset/versioned/fake"
	apps_v1 "k8s.io/api/apps/v1"
	batch_v1 "k8s.io/api/batch/v1"
	core_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/tools/clientcmd/api"
	gatewayapifake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"

	"github.com/kiali/kiali/kubernetes"
)

//// Mock for the K8SClientFactory

type K8SClientFactoryMock struct {
	lock    sync.RWMutex
	Clients map[string]kubernetes.ClientInterface
}

// Constructor
func NewK8SClientFactoryMock(k8s kubernetes.ClientInterface) *K8SClientFactoryMock {
	k8sClientFactory := new(K8SClientFactoryMock)
	k8sClientFactory.Clients = map[string]kubernetes.ClientInterface{kubernetes.HomeClusterName: k8s}
	return k8sClientFactory
}

// Testing specific methods
func (o *K8SClientFactoryMock) SetClients(clients map[string]kubernetes.ClientInterface) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.Clients = clients
}

// Business Methods
func (o *K8SClientFactoryMock) GetClient(authInfo *api.AuthInfo) (kubernetes.ClientInterface, error) {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.Clients[kubernetes.HomeClusterName], nil
}

// Business Methods
func (o *K8SClientFactoryMock) GetClients(authInfo *api.AuthInfo) (map[string]kubernetes.ClientInterface, error) {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.Clients, nil
}

func (o *K8SClientFactoryMock) GetSAClient(cluster string) kubernetes.ClientInterface {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.Clients[cluster]
}

func (o *K8SClientFactoryMock) GetSAClients() map[string]kubernetes.ClientInterface {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.Clients
}

func (o *K8SClientFactoryMock) GetSAHomeClusterClient() kubernetes.ClientInterface {
	o.lock.RLock()
	defer o.lock.RUnlock()
	return o.Clients[kubernetes.HomeClusterName]
}

func (o *K8SClientFactoryMock) GetClusterNames() []string {
	o.lock.RLock()
	defer o.lock.RUnlock()
	clusterNames := make([]string, 0)
	for cn := range o.Clients {
		clusterNames = append(clusterNames, cn)
	}
	return clusterNames
}

/////

type K8SClientMock struct {
	mock.Mock
	istioClientset      *istio_fake.Clientset
	gatewayapiClientSet *gatewayapifake.Clientset
}

// Constructor

func NewK8SClientMock() *K8SClientMock {
	k8s := new(K8SClientMock)
	k8s.On("IsOpenShift").Return(true)
	k8s.On("IsGatewayAPI").Return(false)
	k8s.On("IsIstioAPI").Return(true)
	k8s.On("GetKialiTokenForHomeCluster").Return("")
	k8s.On("GetClusterNames").Return(kubernetes.HomeClusterName)
	return k8s
}

// Business methods

// MockEmptyWorkloads setup the current mock to return empty workloads for every type of workloads (deployment, dc, rs, jobs, etc.)
func (o *K8SClientMock) MockEmptyWorkloads(namespace interface{}) {
	o.On("GetDeployments", namespace).Return([]apps_v1.Deployment{}, nil)
	o.On("GetReplicaSets", namespace).Return([]apps_v1.ReplicaSet{}, nil)
	o.On("GetReplicationControllers", namespace).Return([]core_v1.ReplicationController{}, nil)
	o.On("GetDeploymentConfigs", namespace).Return([]osapps_v1.DeploymentConfig{}, nil)
	o.On("GetStatefulSets", namespace).Return([]apps_v1.StatefulSet{}, nil)
	o.On("GetDaemonSets", namespace).Return([]apps_v1.DaemonSet{}, nil)
	o.On("GetJobs", namespace).Return([]batch_v1.Job{}, nil)
	o.On("GetCronJobs", namespace).Return([]batch_v1.CronJob{}, nil)
}

// MockEmptyWorkload setup the current mock to return an empty workload for every type of workloads (deployment, dc, rs, jobs, etc.)
func (o *K8SClientMock) MockEmptyWorkload(namespace interface{}, workload interface{}) {
	gr := schema.GroupResource{
		Group:    "test-group",
		Resource: "test-resource",
	}
	notfound := errors.NewNotFound(gr, "not found")
	o.On("GetDeployment", namespace, workload).Return(&apps_v1.Deployment{}, notfound)
	o.On("GetStatefulSet", namespace, workload).Return(&apps_v1.StatefulSet{}, notfound)
	o.On("GetDaemonSet", namespace, workload).Return(&apps_v1.DaemonSet{}, notfound)
	o.On("GetDeploymentConfig", namespace, workload).Return(&osapps_v1.DeploymentConfig{}, notfound)
	o.On("GetReplicaSets", namespace).Return([]apps_v1.ReplicaSet{}, nil)
	o.On("GetReplicationControllers", namespace).Return([]core_v1.ReplicationController{}, nil)
	o.On("GetJobs", namespace).Return([]batch_v1.Job{}, nil)
	o.On("GetCronJobs", namespace).Return([]batch_v1.CronJob{}, nil)
}

func (o *K8SClientMock) IsOpenShift() bool {
	args := o.Called()
	return args.Get(0).(bool)
}

func (o *K8SClientMock) IsGatewayAPI() bool {
	args := o.Called()
	return args.Get(0).(bool)
}

func (o *K8SClientMock) IsIstioAPI() bool {
	args := o.Called()
	return args.Get(0).(bool)
}

func (o *K8SClientMock) IsMaistraApi() bool {
	args := o.Called()
	return args.Get(0).(bool)
}

func (o *K8SClientMock) GetServerVersion() (*version.Info, error) {
	args := o.Called()
	return args.Get(0).(*version.Info), args.Error(1)
}

func (o *K8SClientMock) GetToken() string {
	args := o.Called()
	return args.Get(0).(string)
}

func (o *K8SClientMock) GetClusterNames() []string {
	args := o.Called()
	return args.Get(0).([]string)
}

// GetAuthInfo returns the AuthInfo struct for the client
func (o *K8SClientMock) GetAuthInfo() *api.AuthInfo {
	args := o.Called()
	return args.Get(0).(*api.AuthInfo)
}

// GetTokenSubject returns the subject of the authInfo using
// the TokenReview api
func (o *K8SClientMock) GetTokenSubject(authInfo *api.AuthInfo) (string, error) {
	parsedToken, err := jwt.ParseSigned(authInfo.Token)
	if err != nil {
		return authInfo.Token, nil
	}

	var claims map[string]interface{} // generic map to store parsed token
	err = parsedToken.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return authInfo.Token, nil
	}

	if sub, ok := claims["sub"]; ok {
		return sub.(string), nil
	}

	return authInfo.Token, nil
}

func (o *K8SClientMock) MockService(namespace, name string) {
	s := FakeService(namespace, name)
	o.On("GetService", namespace, name).Return(&s, nil)
}

func (o *K8SClientMock) MockServices(namespace string, names []string) {
	services := []core_v1.Service{}
	for _, name := range names {
		services = append(services, FakeService(namespace, name))
	}
	o.On("GetServices", namespace, mock.AnythingOfType("map[string]string")).Return(services, nil)
	o.On("GetDeployments", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]apps_v1.Deployment{}, nil)
}

func FakeService(namespace, name string) core_v1.Service {
	return core_v1.Service{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: core_v1.ServiceSpec{
			ClusterIP: "fromservice",
			Type:      "ClusterIP",
			Selector:  map[string]string{"app": name},
			Ports: []core_v1.ServicePort{
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     3001,
				},
				{
					Name:     "http",
					Protocol: "TCP",
					Port:     3000,
				},
			},
		},
	}
}

func FakePodList() []core_v1.Pod {
	return []core_v1.Pod{
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "reviews-v1",
				Namespace:   "ns",
				Labels:      map[string]string{"app": "reviews", "version": "v1"},
				Annotations: FakeIstioAnnotations(),
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "reviews-v2",
				Namespace:   "ns",
				Labels:      map[string]string{"app": "reviews", "version": "v2"},
				Annotations: FakeIstioAnnotations(),
			},
		},
		{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:        "httpbin-v1",
				Namespace:   "ns",
				Labels:      map[string]string{"app": "httpbin", "version": "v1"},
				Annotations: FakeIstioAnnotations(),
			},
		},
	}
}

func FakeIstioAnnotations() map[string]string {
	return map[string]string{"sidecar.istio.io/status": "{\"version\":\"\",\"initContainers\":[\"istio-init\",\"enable-core-dump\"],\"containers\":[\"istio-proxy\"],\"volumes\":[\"istio-envoy\",\"istio-certs\"]}"}
}

func FakeNamespace(name string) *core_v1.Namespace {
	return &core_v1.Namespace{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: name,
		},
	}
}
