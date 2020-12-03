package sql

import (
	"sync"
)

type Domain struct {
	txCtx   *TxCtx
	schema  string
	name    string
	idMap   map[string]int64
	codeMap map[int64]string
	mux     *sync.Mutex
}

func (o *Domain) load() {
	r := o.txCtx.QuerySql("select id, " + o.name + "code from " + o.schema + "." + o.name)
	defer r.Close()
	for r.Next() {
		var id int64
		var code string
		Scan(r, &id, &code)
		o.codeMap[id] = code
		o.idMap[code] = id
	}
}

func (o *Domain) FindId(code string) *int64 {
	v, ok := o.idMap[code]
	if ok {
		return &v
	} else {
		return nil
	}
}

func (o *Domain) FindCode(id int64) *string {
	v, ok := o.codeMap[id]
	if ok {
		return &v
	} else {
		return nil
	}
}

func NewDomain(txCtx *TxCtx, schema string, name string) *Domain {
	return &Domain{schema: schema, name: name, txCtx: txCtx, idMap: make(map[string]int64), codeMap: make(map[int64]string)}
}

type Domains struct {
	mux       *sync.Mutex
	txCtx     *TxCtx
	schema    string
	domainMap map[string]*Domain
}

func (o *Domains) CheckDomain(name string) *Domain {
	domain := o.domainMap[name]
	if domain == nil {
		o.mux.Lock()
		defer o.mux.Unlock()
		if domain == nil {
			domain = NewDomain(o.txCtx, o.schema, name)
			(*domain).load()
			o.domainMap[name] = domain
		}
	}
	return domain
}

func NewDomais(txCtx *TxCtx, schema string) *Domains {
	return &Domains{txCtx: txCtx, schema: schema, domainMap: map[string]*Domain{}, mux: &sync.Mutex{}}
}
