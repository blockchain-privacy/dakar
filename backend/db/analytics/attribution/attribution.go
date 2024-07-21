package attribution

import (
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/external"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/qrest/gomisc/serror"
	"regexp"
	"time"
)

// AddAttributions adds the given attributions to the database
func AddAttributions(ctx context.Context, c external.Database, attributions []Attribution) error {
	// validate data
	for _, a := range attributions {
		if a.Address.UID == "" || a.Tag == "" || a.Timestamp == "" ||
			(!a.IsPublic && a.User == nil) || (a.IsPublic && a.User != nil) {
			return serror.FromFormat("attribution invalid: %v", a)
		}
	}

	pb, err := json.Marshal(attributions)
	if err != nil {
		return serror.New(err)
	}

	_, err = c.Mutate(ctx, &api.Request{
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
	if err != nil {
		return serror.New(err)
	}

	return nil
}

// GetUserAttributions returns all attributions of a user
func GetUserAttributions(ctx context.Context, c external.Database, userID string) (attributions []FrontendAttribution, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~Attribution.user
				}

				q(func: uid(a)){
					uid
					Attribution.ts
					Attribution.tag
					Attribution.description
					Attribution.source
					Attribution.category
					Attribution.isPublic
					Attribution.address{
						addresshash
					}
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$user": userID})
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Attributions []RequestAttribution `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, attribution := range r.Attributions {
		attributions = append(attributions, attribution.toFrontendAttribution())
	}

	return
}

// DeletePrivateAttribution deletes the given attribution
func DeletePrivateAttribution(ctx context.Context, c external.Database, userID string, attributionUID string) error {
	req := &api.Request{
		Query: `query Q($user:string,$attribution:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~Attribution.user@filter(uid($attribution))
				}
			  }`,
		Vars: map[string]string{"$user": userID, "$attribution": attributionUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(a) * * ."),
		}},
		CommitNow: true,
	}
	resp, err := c.Mutate(ctx, req)
	if err != nil {
		return serror.New(err)
	}

	// check if there was actually something mutated
	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return serror.FromStr("nothing was deleted")
	}

	return nil
}

// DeletePublicAttribution deletes the given public attribution
func DeletePublicAttribution(ctx context.Context, c external.Database, attributionUID string) error {
	req := &api.Request{
		Query: `query Q($attribution:string) {
				a as var(func:uid($attribution))@filter(type(` + DType + ") and eq(Attribution.isPublic,true))}",
		Vars: map[string]string{"$attribution": attributionUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(a) * * ."),
		}},
		CommitNow: true,
	}

	resp, err := c.Mutate(ctx, req)
	if err != nil {
		return serror.New(err)
	}

	// check if there was actually something mutated
	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return serror.FromStr("nothing was deleted")
	}

	return nil
}

// DeleteAllAttributions deletes all attributions of a given user
func DeleteAllAttributions(ctx context.Context, c external.Database, userID string) error {
	req := &api.Request{
		Query: `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~Attribution.user
				}
			  }`,
		Vars: map[string]string{"$user": userID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(a) * * ."),
		}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}

// SearchAttributions returns the attributions that match the query string
func SearchAttributions(ctx context.Context, c external.Database, userID string, searchQuery string) (
	attributions []FrontendAttribution, err error) {
	regex := "/" + regexp.QuoteMeta(searchQuery) + "/i"

	const query = `query Q($user:string,$regex:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~Attribution.user
				}

				pa as var(func:type(` + DType + `))@filter(eq(Attribution.isPublic,true))

				q(func: uid(a, pa), first: 30)@filter(regexp(Attribution.tag,$regex)){
					uid
					Attribution.ts
					Attribution.tag
					Attribution.description
					Attribution.source
					Attribution.category
					Attribution.isPublic
					Attribution.address{
						addresshash
					}
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$user": userID, "$regex": regex})
	if err != nil {
		return nil, serror.New(err)
	}

	var r struct {
		Attributions []RequestAttribution `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, attribution := range r.Attributions {
		attributions = append(attributions, attribution.toFrontendAttribution())
	}

	return
}

// GetAttributionsPerCluster returns all attributions (public and private)
// the user has access to, organized per cluster.
// The returned map is nil if no attributions could be found.
func GetAttributionsPerCluster(c external.Database, userID string, clusterTypes []clustering.ClusterType) (
	attributions map[string][]string, err error) {
	var filter string
	if len(clusterTypes) > 0 {
		for i, ct := range clusterTypes {
			filter += string(ct)

			if i+1 < len(clusterTypes) {
				filter += ","
			}
		}

		filter = "or (eq(Cluster.type," + filter + ") and uid_in(Cluster.user,$user))"
	}

	query := fmt.Sprintf(`query Q($user:string) {
						var(func:uid($user))@filter(type(User)){
							a as ~Attribution.user
						}
						
						pa as var(func:type(`+DType+`))@filter(eq(Attribution.isPublic,true))
						
						q(func: uid(a, pa)){
							uid
							Attribution.address{
								addresshash
								~Cluster.addresses(first:1)@filter(eq(Cluster.type,`+string(clustering.TypeFMI)+`) %s){
									uid
								}
							}
						}
    			     }`, filter)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute, query, map[string]string{"$user": userID})
	if err != nil {
		return
	}

	var r struct {
		Attributions []struct {
			UID     string `json:"uid,omitempty"`
			Address struct {
				Hash    string `json:"addresshash,omitempty"`
				Cluster []struct {
					UID string `json:"uid,omitempty"`
				} `json:"~Cluster.addresses,omitempty"`
			} `json:"Attribution.address,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	if r.Attributions == nil {
		return
	}

	attributions = make(map[string][]string)
	for _, a := range r.Attributions {
		// if address does not have a cluster use address hash as cluster identifier
		if a.Address.Cluster == nil {
			attributions[a.Address.Hash] = append(attributions[a.Address.Hash], a.UID)
		} else {
			attributions[a.Address.Cluster[0].UID] = append(attributions[a.Address.Cluster[0].UID], a.UID)
		}
	}

	return
}
