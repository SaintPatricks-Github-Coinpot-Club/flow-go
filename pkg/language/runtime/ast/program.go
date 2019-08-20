package ast

type Program struct {
	// all declarations, in the order they are defined
	Declarations          []Declaration
	interfaceDeclarations []*InterfaceDeclaration
	structureDeclarations []*StructureDeclaration
	functionDeclarations  []*FunctionDeclaration
	imports               map[ImportLocation]*Program
}

func (p *Program) Accept(visitor Visitor) Repr {
	return visitor.VisitProgram(p)
}

func (p *Program) InterfaceDeclarations() []*InterfaceDeclaration {
	if p.interfaceDeclarations == nil {
		p.interfaceDeclarations = make([]*InterfaceDeclaration, 0)
		for _, declaration := range p.Declarations {
			if interfaceDeclaration, ok := declaration.(*InterfaceDeclaration); ok {
				p.interfaceDeclarations = append(p.interfaceDeclarations, interfaceDeclaration)
			}
		}
	}
	return p.interfaceDeclarations
}

func (p *Program) StructureDeclarations() []*StructureDeclaration {
	if p.structureDeclarations == nil {
		p.structureDeclarations = make([]*StructureDeclaration, 0)
		for _, declaration := range p.Declarations {
			if structureDeclaration, ok := declaration.(*StructureDeclaration); ok {
				p.structureDeclarations = append(p.structureDeclarations, structureDeclaration)
			}
		}
	}
	return p.structureDeclarations
}

func (p *Program) FunctionDeclarations() []*FunctionDeclaration {
	if p.functionDeclarations == nil {
		p.functionDeclarations = make([]*FunctionDeclaration, 0)
		for _, declaration := range p.Declarations {
			if functionDeclaration, ok := declaration.(*FunctionDeclaration); ok {
				p.functionDeclarations = append(p.functionDeclarations, functionDeclaration)
			}
		}
	}
	return p.functionDeclarations
}

func (p *Program) Imports() map[ImportLocation]*Program {
	if p.imports == nil {
		p.imports = make(map[ImportLocation]*Program)
		for _, declaration := range p.Declarations {
			if importDeclaration, ok := declaration.(*ImportDeclaration); ok {
				p.imports[importDeclaration.Location] = nil
			}
		}
	}
	return p.imports
}
