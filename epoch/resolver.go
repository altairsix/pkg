package epoch

import "time"

// Resolver provides a resolver usable by github.com/neelance/graphql-go
//
//	type Epoch {
//		date: String!
//		time: String!
//		secs: Int!
//		ago: String!
//	}
//
type Resolver struct {
	em Millis
}

func (r *Resolver) Date() string {
	return r.em.Format("1/2/2006")
}

func (r *Resolver) Time() string {
	return r.em.Format(time.Kitchen)
}

func (r *Resolver) Secs() int32 {
	return int32(r.em.Int64() / 1000)
}

func (r *Resolver) Ago() string {
	return (Now() - r.em).Ago().String
}
