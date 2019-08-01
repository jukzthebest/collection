package collection

import (
	"fmt"
	"github.com/shopspring/decimal"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type MapArrayCollection struct {
	value []map[string]interface{}
	BaseCollection
}

func (c MapArrayCollection) Value() interface{} {
	return c.value
}

func (c MapArrayCollection) Sum(key ...string) decimal.Decimal {
	var sum = decimal.New(0, 0)

	for i := 0; i < len(c.value); i++ {
		sum = sum.Add(nd(c.value[i][key[0]]))
	}

	return sum
}

func (c MapArrayCollection) Median(key ...string) decimal.Decimal {
	var f = make([]decimal.Decimal, len(c.value))
	for i := 0; i < len(c.value); i++ {
		f = append(f, nd(c.value[i][key[0]]))
	}
	qsort(f, true)
	return f[len(f)/2].Add(f[len(f)/2-1]).Div(nd(2))
}

func (c MapArrayCollection) Min(key ...string) decimal.Decimal {

	var (
		smallest = decimal.New(0, 0)
		number   decimal.Decimal
	)

	for i := 0; i < len(c.value); i++ {
		number = nd(c.value[i][key[0]])
		if i == 0 {
			smallest = number
			continue
		}
		if smallest.GreaterThan(number) {
			smallest = number
		}
	}

	return smallest
}

func (c MapArrayCollection) Max(key ...string) decimal.Decimal {

	var (
		biggest = decimal.New(0, 0)
		number  decimal.Decimal
	)

	for i := 0; i < len(c.value); i++ {
		number = nd(c.value[i][key[0]])
		if i == 0 {
			biggest = number
			continue
		}
		if biggest.LessThan(number) {
			biggest = number
		}
	}

	return biggest
}

func (c MapArrayCollection) Pluck(key string) Collection {
	var s = make([]interface{}, 0)
	for i := 0; i < len(c.value); i++ {
		s = append(s, c.value[i][key])
	}
	return Collect(s)
}

func (c MapArrayCollection) Prepend(values ...interface{}) Collection {

	var d MapArrayCollection

	var n = make([]map[string]interface{}, len(c.value))
	copy(n, c.value)

	d.value = append([]map[string]interface{}{values[0].(map[string]interface{})}, n...)
	d.length = len(d.value)

	return d
}

func (c MapArrayCollection) Only(keys []string) Collection {
	var d MapArrayCollection

	var ma = make([]map[string]interface{}, 0)
	for _, k := range keys {
		m := make(map[string]interface{}, 0)
		for _, v := range c.value {
			m[k] = v[k]
		}
		ma = append(ma, m)
	}
	d.value = ma
	d.length = len(ma)

	return d
}

func (c MapArrayCollection) Splice(index ...int) Collection {

	if len(index) == 1 {
		var n = make([]map[string]interface{}, len(c.value))
		copy(n, c.value)
		n = n[index[0]:]

		return MapArrayCollection{n, BaseCollection{length: len(n)}}
	} else if len(index) > 1 {
		var n = make([]map[string]interface{}, len(c.value))
		copy(n, c.value)
		n = n[index[0] : index[0]+index[1]]

		return MapArrayCollection{n, BaseCollection{length: len(n)}}
	} else {
		panic("invalid argument")
	}
}

func (c MapArrayCollection) Take(num int) Collection {
	var d MapArrayCollection
	if num > c.length {
		panic("not enough elements to take")
	}

	if num >= 0 {
		d.value = c.value[:num]
		d.length = num
	} else {
		d.value = c.value[len(c.value)+num:]
		d.length = 0 - num
	}

	return d
}

func (c MapArrayCollection) All() []interface{} {
	s := make([]interface{}, len(c.value))
	for i := 0; i < len(c.value); i++ {
		s[i] = c.value[i]
	}

	return s
}

func (c MapArrayCollection) Mode(key ...string) []interface{} {
	valueCount := make(map[interface{}]int)
	for i := 0; i < c.length; i++ {
		if v, ok := c.value[i][key[0]]; ok {
			valueCount[v]++
		}
	}

	maxCount := 0
	maxValue := make([]interface{}, len(valueCount))
	for v, c := range valueCount {
		switch {
		case c < maxCount:
			continue
		case c == maxCount:
			maxValue = append(maxValue, v)
		case c > maxCount:
			maxValue = append([]interface{}{}, v)
			maxCount = c
		}
	}
	return maxValue
}

func (c MapArrayCollection) ToMapArray() []map[string]interface{} {
	return c.value
}

func (c MapArrayCollection) Chunk(num int) MultiDimensionalArrayCollection {
	var d MultiDimensionalArrayCollection
	d.length = c.length/num + 1
	d.value = make([][]interface{}, d.length)

	count := 0
	for i := 1; i <= c.length; i++ {
		switch {
		case i == c.length:
			if i%num == 0 {
				d.value[count] = c.All()[i-num:]
				d.value = d.value[:d.length-1]
			} else {
				d.value[count] = c.All()[i-i%num:]
			}
		case i%num != 0 || i < num:
			continue
		default:
			d.value[count] = c.All()[i-num : i]
			count++
		}
	}

	return d
}

func (c MapArrayCollection) Concat(value interface{}) Collection {
	return MapArrayCollection{
		value:          append(c.value, value.([]map[string]interface{})...),
		BaseCollection: BaseCollection{length: c.length + len(value.([]map[string]interface{}))},
	}
}

func (c MapArrayCollection) Contains(value interface{}, callback ...interface{}) bool {
	if len(callback) != 0 {
		return callback[0].(func() bool)()
	}

	t := fmt.Sprintf("%T", c.value)
	switch {
	case t == "[]map[string]string":
		for _, m := range c.value {
			if parseContainsParam(m, intToString(value)) {
				return true
			}
		}
		return false
	default:
		for _, m := range c.value {
			if parseContainsParam(m, value) {
				return true
			}
		}
		return false
	}
}

func parseContainsParam(m map[string]interface{}, value interface{}) bool {
	switch value.(type) {
	case map[string]interface{}:
		return containsKeyValue(m, value.(map[string]interface{}))
	default:
		return containsValue(m, value)
	}
}

func intToString(value interface{}) interface{} {
	switch value.(type) {
	case int:
		return strconv.Itoa(value.(int))
	case int64:
		return strconv.FormatInt(value.(int64), 10)
	default:
		return value
	}
}

func containsValue(m interface{}, value interface{}) bool {
	switch m.(type) {
	case map[string]interface{}:
		for _, v := range m.(map[string]interface{}) {
			if v == value {
				return true
			}
		}
		return false
	case []decimal.Decimal:
		for _, v := range m.([]decimal.Decimal) {
			if v.Equal(nd(value)) {
				return true
			}
		}
		return false
	case []string:
		for _, v := range m.([]string) {
			if v == value {
				return true
			}
		}
		return false
	default:
		panic("wrong type")
	}
}

func containsKeyValue(m map[string]interface{}, value map[string]interface{}) bool {
	for k, v := range value {
		if _, ok := m[k]; !ok && m[k] != v {
			return false
		}
	}

	return true
}

func (c MapArrayCollection) ContainsStrict(value interface{}, callback ...interface{}) bool {
	if len(callback) != 0 {
		return callback[0].(func() bool)()
	}

	for _, m := range c.value {
		if parseContainsParam(m, value) {
			return true
		}
	}

	return false
}

func (c MapArrayCollection) CrossJoin(array ...[]interface{}) MultiDimensionalArrayCollection {
	var d MultiDimensionalArrayCollection

	// A two-dimensional-slice's initial
	length := len(c.value)
	for _, s := range array {
		length *= len(s)
	}
	value := make([][]interface{}, length)
	for i := range value {
		value[i] = make([]interface{}, len(array)+1)
	}

	offset := length / c.length
	for i := 0; i < length; i++ {
		value[i][0] = c.value[i/offset]
	}
	assignmentToValue(value, array, length, 1, 0, offset)

	d.value = value
	d.length = length
	return d
}

// vl: length of value
// ai: index of array
// si: index of value's sub-array
func assignmentToValue(value, array [][]interface{}, vl, si, ai, preOffset int) {
	offset := preOffset / len(array[ai])
	times := 0

	for i := 0; i < vl; i++ {
		if i >= preOffset && i%preOffset == 0 {
			times++
		}
		value[i][si] = array[ai][(i-preOffset*times)/offset]
	}

	if ai < len(array)-1 {
		assignmentToValue(value, array, vl, si+1, ai+1, offset)
	}
}

func (c MapArrayCollection) Dd() {
	dd(c)
}

func (c MapArrayCollection) Dump() {
	dump(c)
}

func (c MapArrayCollection) Every(cb CB) bool {
	for key, value := range c.value {
		if !cb(key, value) {
			return false
		}
	}
	return true
}

func (c MapArrayCollection) Filter(cb CB) Collection {
	var d = make([]map[string]interface{}, 0)
	copy(d, c.value)
	for key, value := range c.value {
		if !cb(key, value) {
			d = append(d[:key], d[key+1:]...)
		}
	}
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) First(cbs ...CB) interface{} {
	if len(cbs) > 0 {
		for key, value := range c.value {
			if cbs[0](key, value) {
				return value
			}
		}
		return nil
	} else {
		if len(c.value) > 0 {
			return c.value[0]
		} else {
			return nil
		}
	}
}

func (c MapArrayCollection) FirstWhere(key string, values ...interface{}) map[string]interface{} {
	if len(values) < 1 {
		for _, value := range c.value {
			if isTrue(value[key]) {
				return value
			}
		}
	} else if len(values) < 2 {
		for _, value := range c.value {
			if value[key] == values[0] {
				return value
			}
		}
	} else {
		switch values[0].(string) {
		case ">":
			for _, value := range c.value {
				if nd(value[key]).GreaterThan(nd(values[0])) {
					return value
				}
			}
		case ">=":
			for _, value := range c.value {
				if nd(value[key]).GreaterThanOrEqual(nd(values[0])) {
					return value
				}
			}
		case "<":
			for _, value := range c.value {
				if nd(value[key]).LessThan(nd(values[0])) {
					return value
				}
			}
		case "<=":
			for _, value := range c.value {
				if nd(value[key]).LessThanOrEqual(nd(values[0])) {
					return value
				}
			}
		case "=":
			for _, value := range c.value {
				if value[key] == values[0] {
					return value
				}
			}
		}
	}
	return map[string]interface{}{}
}

func (c MapArrayCollection) GroupBy(k string) Collection {
	var d = make(map[string]interface{}, 0)
	for _, value := range c.value {
		for kk, vv := range value {
			if kk == k {
				vvKey := fmt.Sprintf("%v", vv)
				if _, ok := d[vvKey]; ok {
					am := d[vvKey].([]map[string]interface{})
					am = append(am, value)
					d[vvKey] = am
				} else {
					d[vvKey] = []map[string]interface{}{value}
				}
			}
		}
	}
	return MapCollection{
		value: d,
	}
}

func (c MapArrayCollection) Implode(key string, delimiter string) string {
	var res = ""
	for _, value := range c.value {
		for kk, vv := range value {
			if kk == key {
				res += fmt.Sprintf("%v", vv) + delimiter
			}
		}
	}
	return res[:len(res)-1]
}

func (c MapArrayCollection) IsEmpty() bool {
	return len(c.value) == 0
}

func (c MapArrayCollection) IsNotEmpty() bool {
	return len(c.value) != 0
}

func (c MapArrayCollection) KeyBy(v interface{}) Collection {
	var d = make(map[string]interface{}, 0)
	if k, ok := v.(string); ok {
		for _, value := range c.value {
			for kk, vv := range value {
				if kk == k {
					d[fmt.Sprintf("%v", vv)] = []map[string]interface{}{value}
				}
			}
		}
	} else {
		vb := v.(FilterFun)
		for _, value := range c.value {
			for kk, vv := range value {
				if kk == k {
					d[fmt.Sprintf("%v", vb(vv))] = []map[string]interface{}{value}
				}
			}
		}
	}
	return MapCollection{
		value: d,
	}
}

func (c MapArrayCollection) Last(cbs ...CB) interface{} {
	if len(cbs) > 0 {
		var last interface{}
		for key, value := range c.value {
			if cbs[0](key, value) {
				last = value
			}
		}
		return last
	} else {
		if len(c.value) > 0 {
			return c.value[len(c.value)-1]
		} else {
			return nil
		}
	}
}

func (c MapArrayCollection) MapToGroups(cb MapCB) Collection {
	var d = make(map[string]interface{}, 0)
	for _, value := range c.value {
		nk, nv := cb(value)
		if _, ok := d[nk]; ok {
			am := d[nk].([]interface{})
			am = append(am, nv)
			d[nk] = am
		} else {
			d[nk] = []interface{}{nv}
		}
	}
	return MapCollection{
		value: d,
	}
}

func (c MapArrayCollection) MapWithKeys(cb MapCB) Collection {
	var d = make(map[string]interface{}, 0)
	for _, value := range c.value {
		nk, nv := cb(value)
		d[nk] = nv
	}
	return MapCollection{
		value: d,
	}
}

func (c MapArrayCollection) Partition(cb PartCB) (Collection, Collection) {
	var d1 = make([]map[string]interface{}, 0)
	var d2 = make([]map[string]interface{}, 0)

	for i := 0; i < len(c.value); i++ {
		if cb(i) {
			d1 = append(d1, c.value[i])
		} else {
			d2 = append(d2, c.value[i])
		}
	}

	return MapArrayCollection{
		value: d1,
	}, MapArrayCollection{
		value: d2,
	}
}

func (c MapArrayCollection) Pop() interface{} {
	last := c.value[len(c.value)-1]
	c.value = c.value[:len(c.value)-1]
	return last
}

func (c MapArrayCollection) Push(v interface{}) Collection {
	var d = make([]map[string]interface{}, len(c.value)+1)
	for i := 0; i < len(d); i++ {
		if i < len(c.value) {
			d[i] = c.value[i]
		} else {
			d[i] = v.(map[string]interface{})
		}
	}

	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Random(num ...int) Collection {
	if len(num) == 0 {
		return BaseCollection{
			value: c.value[rand.Intn(len(c.value))],
		}
	} else {
		if num[0] > len(c.value) {
			panic("wrong num")
		}
		var d = make([]map[string]interface{}, len(c.value))
		copy(d, c.value)
		for i := 0; i < len(c.value)-num[0]; i++ {
			index := rand.Intn(len(d))
			d = append(d[:index], d[index+1:]...)
		}
		return MapArrayCollection{
			value: d,
		}
	}
}

func (c MapArrayCollection) Reduce(cb ReduceCB) interface{} {
	var res interface{}

	for i := 0; i < len(c.value); i++ {
		res = cb(res, c.value[i])
	}

	return res
}

func (c MapArrayCollection) Reject(cb CB) Collection {
	var d = make([]map[string]interface{}, 0)
	for key, value := range c.value {
		if !cb(key, value) {
			d = append(d, value)
		}
	}
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Reverse() Collection {
	var d = make([]map[string]interface{}, len(c.value))
	j := 0
	for i := len(c.value) - 1; i > -1; i-- {
		d[j] = c.value[i]
		j++
	}
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Search(v interface{}) int {
	cb := v.(CB)
	for i := 0; i < len(c.value); i++ {
		if cb(i, c.value[i]) {
			return i
		}
	}
	return -1
}

func (c MapArrayCollection) Shift() Collection {
	var d = make([]map[string]interface{}, len(c.value))
	copy(d, c.value)
	d = d[1:]
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Shuffle() Collection {
	var d = make([]map[string]interface{}, len(c.value))
	copy(d, c.value)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(c.value), func(i, j int) { d[i], d[j] = d[j], d[i] })
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Slice(keys ...int) Collection {
	var d = make([]map[string]interface{}, len(c.value))
	copy(d, c.value)
	if len(keys) == 1 {
		return MapArrayCollection{
			value: d[keys[0]:],
		}
	} else {
		return MapArrayCollection{
			value: d[keys[0] : keys[0]+keys[1]],
		}
	}
}

func (c MapArrayCollection) Split(num int) Collection {
	var d = make([][]interface{}, math.Ceil(float64(len(c.value))/float64(num)))

	j := 0
	for i := 0; i < len(c.value); i++ {
		if i%num == 0 {
			if i+num <= len(c.value) {
				d[j] = make([]interface{}, num)
			} else {
				d[j] = make([]interface{}, len(c.value)-i)
			}
			d[j][i%num] = c.value[i]
			j++
		} else {
			d[j][i%num] = c.value[i]
		}
	}

	return MultiDimensionalArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) WhereIn(key string, in []interface{}) Collection {
	var d = make([]map[string]interface{}, 0)
	for i := 0; i < len(c.value); i++ {
		for j := 0; j < len(in); j++ {
			if c.value[i][key] == in[j] {
				d = append(d, copyMap(c.value[i]))
				break
			}
		}
	}
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) WhereNotIn(key string, in []interface{}) Collection {
	var d = make([]map[string]interface{}, 0)
	for i := 0; i < len(c.value); i++ {
		for j := 0; j < len(in); j++ {
			isIn := false
			if c.value[i][key] == in[j] {
				isIn = true
				break
			}
			if !isIn {
				d = append(d, copyMap(c.value[i]))
			}
		}
	}
	return MapArrayCollection{
		value: d,
	}
}

func (c MapArrayCollection) Where(key string, values ...interface{}) Collection {
	var d = make([]map[string]interface{}, 0)
	if len(values) < 1 {
		for _, value := range c.value {
			if isTrue(value[key]) {
				d = append(d, copyMap(value))
			}
		}
	} else if len(values) < 2 {
		for _, value := range c.value {
			if value[key] == values[0] {
				d = append(d, copyMap(value))
			}
		}
	} else {
		switch values[0].(string) {
		case ">":
			for _, value := range c.value {
				if nd(value[key]).GreaterThan(nd(values[0])) {
					d = append(d, copyMap(value))
				}
			}
		case ">=":
			for _, value := range c.value {
				if nd(value[key]).GreaterThanOrEqual(nd(values[0])) {
					d = append(d, copyMap(value))
				}
			}
		case "<":
			for _, value := range c.value {
				if nd(value[key]).LessThan(nd(values[0])) {
					d = append(d, copyMap(value))
				}
			}
		case "<=":
			for _, value := range c.value {
				if nd(value[key]).LessThanOrEqual(nd(values[0])) {
					d = append(d, copyMap(value))
				}
			}
		case "=":
			for _, value := range c.value {
				if value[key] == values[0] {
					d = append(d, copyMap(value))
				}
			}
		}
	}
	return MapArrayCollection{
		value: d,
	}
}