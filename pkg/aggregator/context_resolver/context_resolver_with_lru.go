package context_resolver

import (
	"github.com/DataDog/datadog-agent/pkg/aggregator/ckey"
	"github.com/DataDog/datadog-agent/pkg/metrics"
)

// ContextResolver allows tracking and expiring contexts
type contextResolverWithLRU struct {
	contextResolverBase
	cache *contextResolverLru
	resolver ContextResolver
}

func NewcontextResolverWithLRU(resolver ContextResolver, cacheSize int) *contextResolverWithLRU {
	return &contextResolverWithLRU{
		contextResolverBase: newContextResolverBase(),
		cache: NewContextResolverLru(cacheSize),
		resolver: resolver,
	}
}

// TrackContext returns the contextKey associated with the context of the metricSample and tracks that context
func (cr *contextResolverWithLRU) TrackContext(metricSampleContext metrics.MetricSampleContext) ckey.ContextKey {
	// There is room for optimization here are both methods are doing similar things.
	var keyCache = cr.cache.TrackContext(metricSampleContext)
	var keyResolver = cr.resolver.TrackContext(metricSampleContext)
	if keyCache != keyResolver {
		panic("Misconfigured resolvers are returning different keys")
	}
	return keyResolver
}

// Get gets a context from its key
func (cr *contextResolverWithLRU) Get(key ckey.ContextKey) (*Context, bool) {
	ctx, found := cr.cache.Get(key)
	if !found {
		ctx, found = cr.resolver.Get(key)
	}
	return ctx, found
}

// Size returns the number of contexts in the resolver
func (cr *contextResolverWithLRU) Size() int {
	return cr.resolver.Size()
}

func (cr *contextResolverWithLRU) removeKeys(expiredContextKeys []ckey.ContextKey) {
	cr.cache.removeKeys(expiredContextKeys)
	cr.resolver.removeKeys(expiredContextKeys)
}


func (cr *contextResolverWithLRU) Clear() {
	cr.cache.Clear()
	cr.resolver.Clear()
}

func (cr *contextResolverWithLRU) Close() {
	cr.cache.Close()
	cr.resolver.Close()
}
