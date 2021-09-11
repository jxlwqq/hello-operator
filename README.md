# hello-operator

### 前置条件

* 安装 Docker Desktop，并启动内置的 Kubernetes 集群
* 注册一个 [hub.docker.com](https://hub.docker.com/) 账户，需要将本地构建好的镜像推送至公开仓库中
* 安装 operator SDK CLI: `brew install operator-sdk`
* 安装 Go: `brew install go`

本示例推荐的依赖版本：

* Docker Desktop: >= 4.0.0
* Kubernetes: >= 1.21.4
* Operator-SDK: >= 1.11.0
* Go: >= 1.17

### 创建项目

使用 Operator SDK CLI 创建名为 hello-operator 的项目。

```shell
mkdir -p $HOME/projects/hello-operator
cd $HOME/projects/hello-operator
go env -w GOPROXY=https://goproxy.cn,direct

```shell

operator-sdk init --domain=jxlwqq.github.io \
--repo=github.com/jxlwqq/hello-operator \
--skip-go-version-check
```


### 创建 API 和控制器

使用 Operator SDK CLI 创建自定义资源定义（CRD）API 和控制器。

运行以下命令创建带有组 app、版本 v1alpha1 和种类 Hello 的 API：

```shell
operator-sdk create api \
--resource=true \
--controller=true \
--group=app \
--version=v1alpha1 \
--kind=Hello
```


定义 Hello 自定义资源（CR）的 API。

修改 api/v1alpha1/hello_types.go 中的 Go 类型定义，使其具有以下 spec 和 status

```go
type HellloSpec struct {
	Size int32 `json:"size"`
	Version string `json:"version"`
}
```

为资源类型更新生成的代码：
```shell
make generate
```

运行以下命令以生成和更新 CRD 清单：
```shell
make manifests
```

运行以下命令以生成和更新 CRD 清单：
```shell
make manifests
```

#### 实现控制器

> 由于逻辑较为复杂，代码较为庞大，所以无法在此全部展示，完整的操作器代码请参见 controllers 目录。
在本例中，将生成的控制器文件 controllers/hello_controller.go 替换为以下示例实现：
```go
/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/jxlwqq/hello-operator/api/v1alpha1"
)

// HelloReconciler reconciles a Hello object
type HelloReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=app.jxlwqq.github.io,resources=hellos/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Hello object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *HelloReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Reconciling Hello")

	hello := &appv1alpha1.Hello{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, hello)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	var result *ctrl.Result
	reqLogger.Info("---Frontend Deployment---")
	result, err = r.ensureDeployment(r.frontendDeployment(hello))
	if result != nil {
		return *result, err
	}

	reqLogger.Info("---Frontend Service---")
	result, err = r.ensureService(r.frontendService(hello))
	if result != nil {
		return *result, err
	}

	reqLogger.Info("---Frontend Change Handler---")
	result, err = r.handleFrontendChanges(hello)
	if result != nil {
		return *result, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.Hello{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
```

运行以下命令以生成和更新 CRD 清单：
```shell
make manifests
```


#### 运行 Operator

捆绑 Operator，并使用 Operator Lifecycle Manager（OLM）在集群中部署。

修改 Makefile 中 IMAGE_TAG_BASE 和 IMG：

```makefile
IMAGE_TAG_BASE ?= docker.io/jxlwqq/hello-operator
IMG ?= $(IMAGE_TAG_BASE):latest
```

构建镜像：

```shell
make docker-build
```

将镜像推送到镜像仓库：
```shell
make docker-push
```

成功后访问：https://hub.docker.com/r/jxlwqq/hello-operator

运行 make bundle 命令创建 Operator 捆绑包清单，并依次填入名称、作者等必要信息:
```shell
make bundle
```

构建捆绑包镜像：
```shell
make bundle-build
```

推送捆绑包镜像：
```shell
make bundle-push
```

成功后访问：https://hub.docker.com/r/jxlwqq/hello-operator-bundle



使用 Operator Lifecycle Manager 部署 Operator:

```shell
# 切换至本地集群
kubectl config use-context docker-desktop
# 安装 olm
operator-sdk olm install
# 使用 Operator SDK 中的 OLM 集成在集群中运行 Operator
operator-sdk run bundle docker.io/jxlwqq/hello-operator-bundle:v0.0.1
```

### 创建自定义资源

编辑 config/samples/app_v1alpha1_hello.yaml 上的 Hello CR 清单示例，使其包含以下规格：

```yaml
apiVersion: app.jxlwqq.github.io/v1alpha1
kind: Hello
metadata:
  name: hello-sample
spec:
  # Add fields here
  size: 2
  version: "1.9"

```

创建 CR：
```shell
kubectl apply -f config/samples/app_v1alpha1_hello.yaml
```

查看 Pod：
```shell
NAME                     READY   STATUS    RESTARTS   AGE
hello-5cdd78d697-8jz8k   1/1     Running   0          9s
hello-5cdd78d697-rmrhd   1/1     Running   0          9s
```

查看 Service：
```shell
NAME         TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
hello-svc    NodePort    10.98.205.249   <none>        8080:30691/TCP   18s
kubernetes   ClusterIP   10.96.0.1       <none>        443/TCP          3h9m
```

浏览器访问：http://localhost:30691

网页上会显示出 Hello world! 的欢迎页面。

更新 CR：

```shell
# 修改副本数和 Hello 版本
kubectl patch hello hello-sample -p '{"spec":{"size": 3, "version": "1.10"}}' --type=merge
```

查看 Pod：
```shell
NAME                     READY   STATUS    RESTARTS   AGE
hello-5bcdc7b75f-p4ktw   1/1     Running   0          7s
hello-5bcdc7b75f-p5tkx   1/1     Running   0          7s
hello-5bcdc7b75f-qswxd   1/1     Running   0          7s
```

#### 做好清理

```shell
operator-sdk cleanup hello-operator
operator-sdk olm uninstall
```