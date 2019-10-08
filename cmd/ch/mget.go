package main

import (
	"github.com/coyove/iis/cmd/ch/id"
	"github.com/etcd-io/bbolt"
)

func (m *Manager) get(main *bbolt.Bucket, id []byte) (a *Article) {
	sid := string(id)
	if v, ok := m.cache.Get(sid); ok {
		return v.(*Article)
	}

	a = &Article{}
	a.unmarshal(main.Get(id))

	if a.ID != nil {
		m.cache.Add(sid, a)
	}
	return a
}

func (m *Manager) mget(main *bbolt.Bucket, tl bool, res [][2][]byte) (a []*Article) {
	for _, r := range res {
		if hdr := r[0][0]; tl {
			// If in timeline mode, we accept two headers:
			if hdr != id.HeaderTimeline && hdr != id.HeaderAnnounce {
				continue
			}
		} else {
			// If in author-tag mode, we accept one header:
			if hdr != id.HeaderAuthorTag {
				continue
			}
		}

		if p := m.get(main, r[1]); p.ID != nil {
			a = append(a, p)
		}
	}
	return
}

func (m *Manager) mgetReplies(parent []byte, start, end int) (a []*Article) {
	m.db.View(func(tx *bbolt.Tx) error {
		main := tx.Bucket(bkPost)

		for i := start; i < end; i++ {
			if i <= 0 || i >= 128*128 {
				continue
			}

			pid := id.ParseID(parent)
			pid.RIndexAppend(int16(i))
			pid.SetHeader(id.HeaderReply)

			p := m.get(main, pid.Marshal())
			if p.ID == nil {
				p.NotFound = true
				p.Index = int64(i)
			}
			a = append(a, p)
		}
		return nil
	})
	return
}
