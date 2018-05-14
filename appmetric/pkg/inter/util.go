package inter

func GenerateFakeMetrics() []*EntityMetric {
	result := []*EntityMetric{}

	ip1 := "10.0.2.3"
	em := NewEntityMetric(ip1, ApplicationType)
	em.SetLabel("name", "default/curl-1xfj")
	em.SetLabel("ip", ip1)

	em.SetMetric(Latency, 133.2)
	em.SetMetric(TPS, 12)
	result = append(result, em)

	ip2 := "10.0.3.2"
	em2 := NewEntityMetric(ip2, ApplicationType)
	em2.SetLabel("name", "istio/music-ftaf2")
	em2.SetLabel("ip", ip2)

	em2.SetMetric(Latency, 13.2)
	em2.SetMetric(TPS, 10)
	result = append(result, em2)

	return result
}
