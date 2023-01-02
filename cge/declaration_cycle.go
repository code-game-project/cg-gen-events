package cge

import (
	"fmt"
	"strings"
)

type declCycleObj struct {
	o        Object
	hadError bool
}

type declarationCycleDetector struct {
	parser  *parser
	objects map[string]*declCycleObj

	stack []*declCycleObj
}

func (p *parser) detectDeclarationCycles() {
	detector := &declarationCycleDetector{
		parser: p,
		stack:  make([]*declCycleObj, 0, 5),
	}

	detector.objects = make(map[string]*declCycleObj, len(detector.parser.objects))
	for _, o := range detector.parser.objects {
		detector.objects[o.Name.Lexeme] = &declCycleObj{
			o: o,
		}
	}

	detector.find()
}

func (d *declarationCycleDetector) find() {
	for _, o := range d.objects {
		d.check(o)
	}
}

func (d *declarationCycleDetector) check(obj *declCycleObj) {
	if obj.o.Type != TYPE {
		return
	}

	if i := d.findInStack(obj); i != -1 {
		if !obj.hadError {
			obj.hadError = true
			d.pushToStack(obj)
			names := make([]string, len(d.stack)-i)
			for j, o := range d.stack[i:len(d.stack)] {
				names[j] = o.o.Name.Lexeme
			}
			d.parser.errors = append(d.parser.errors, d.parser.newErrorAt(fmt.Sprintf("Declaration cycle detected: %s", strings.Join(names, "->")), obj.o.Name, false))
			d.popFromStack()
		}
		return
	}

	d.pushToStack(obj)

	for _, p := range obj.o.Properties {
		o, ok := d.objects[p.Type.Token.Lexeme]
		if ok {
			d.check(o)
		}
	}

	d.popFromStack()
}

func (d *declarationCycleDetector) pushToStack(obj *declCycleObj) {
	d.stack = append(d.stack, obj)
}

func (d *declarationCycleDetector) popFromStack() *declCycleObj {
	o := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]
	return o
}

func (d *declarationCycleDetector) findInStack(obj *declCycleObj) int {
	for i, o := range d.stack {
		if o == obj {
			return i
		}
	}
	return -1
}
