# addOn
Add other kinds of entity getter: Get entities and their metrics from different kinds of Prometheus exporters.
Currently, [Istio exporter](https://istio.io/docs/reference/config/adapters/prometheus.html) and [Redis exporter](https://github.com/oliver006/redis_exporter) are supported.


# How to add support for other kinds of Prometheus exporters

#### Step1 Implement the `EntityMetricGetter` interface
To get entities from other kinds of exporters, implement `EntityMetricGetter`:
```golang
type EntityMetricGetter interface {
	GetEntityMetric(client *prometheus.RestClient) ([]*inter.EntityMetric, error)
	Name() string
}
```

The `Name() string` function needs to return a unique string from other entity getter instances.

The input of `GetEntityMetric()` is a [Prometheus REST client](https://github.com/songbinliu/xfire/blob/1667ae6ade0c27b7c30c514574b9bd3e886b5258/pkg/prometheus/prometheus_client.go#L23);
and its output is a list of [`EntityMetric`](https://github.com/songbinliu/appMetric/blob/020e76fcd2a261fbbb4429e6109013db72ff1b4f/pkg/inter/types.go#L3).


#### Step2 Add the new addon to the Factory
Add the new implemented addon to the [GetterFactory](https://github.com/songbinliu/appMetric/blob/020e76fcd2a261fbbb4429e6109013db72ff1b4f/pkg/addon/factory.go#L21).

```golang
func (f *GetterFactory) CreateEntityGetter(category, name string) (alligator.EntityMetricGetter, error) {
	switch category {
	case RedisGetterCategory:
		return NewRedisEntityGetter(name), nil
	case IstioGetterCategory:
		g := newIstioEntityGetter(name)
		g.SetType(false)
		return g, nil
	case IstioVAppGetterCategory:
		g := newIstioEntityGetter(name)
		g.SetType(true)
		return g, nil
	}

	return nil, fmt.Errorf("Unknown category: %v", category)
}
```
