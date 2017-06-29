package epoch

import "time"

// Resolver provides a resolver usable by github.com/neelance/graphql-go
//
//	type Epoch {
//		date: String!
//		time: String!
//		value: Int!
//		ago: String!
//	}
type Resolver struct {
	em Millis
}

func (r *Resolver) Date() string {
	return r.em.Format("1/2/2006")
}

func (r *Resolver) Time() string {
	return r.em.Format(time.Kitchen)
}

func (r *Resolver) Value() int64 {
	return r.em.Int64()
}

func (r *Resolver) Ago() string {
	return (Now() - r.em).Ago().String
}
