#
#   make_ast.py
#   ~~~~~~~~~~~
#
#   generates ast nodes for the toe language.
#

import sys
from subprocess import Popen, PIPE


PACKAGE_TEMPLATE = """\
// generated by tool/make_ast.py, do not modify!
package parser

import "toe/lexer"

{decls}
"""

STRUCT_TEMPLATE = """\
type {name} struct {{
{fields}
}}

{functions}
"""

METH_TEMPLATE = "func (node *{node_type}) {method} {{ {body} }}"
CONS_TEMPLATE = "func new{node_type}({args}) *{node_type} {{\nreturn &{node_type}{{\n{props}\n}}\n}}"


class Struct:
    """
    An AST-node to be generated. The arguments are:

    1. `name`: is the name of the struct (e.g. Identifier, Module, ...),
    2. `fields`: contains the mandatory fields when creating the node,
    3. `extra_fields`: contains extra fields (e.g. annotations),
    4. `methods`: a list of methods -- see METH_TEMPLATE.
    """

    def __init__(self, name, fields, extra_fields=(), methods=()):
        self.name = name
        self.fields = fields
        self.extra_fields = list(extra_fields)
        self.methods = list(methods)

    def generate(self):
        methods = [METH_TEMPLATE.format_map({'node_type': self.name, **method})
                   for method in self.methods]
        constructor = CONS_TEMPLATE.format(
            node_type=self.name,
            args=', '.join(self.fields),
            props='\n'.join('{prop}:{prop},'.format(prop=x.split(' ')[0])
                            for x in self.fields),
        )
        return STRUCT_TEMPLATE.format(
            name=self.name,
            fields='\n'.join(self.fields + self.extra_fields),
            functions='\n'.join([constructor] + methods)
        )


def token_method(field):
    return {"method": "Tok() lexer.Token",
            "body": f"return node.{field}"}


def generate(*, stmts, exprs):
    seen = set()
    structs = []

    for struct in stmts:
        assert struct.name not in seen
        struct.methods.append({"method": "node()", "body": ""})
        struct.methods.append({"method": "stmt()", "body": ""})
        structs.append(struct)
        seen.add(struct.name)

    for struct in exprs:
        assert struct.name not in seen
        struct.methods.append({"method": "node()", "body": ""})
        struct.methods.append({"method": "expr()", "body": ""})
        structs.append(struct)
        seen.add(struct.name)

    package = PACKAGE_TEMPLATE.format(
        decls='\n'.join(s.generate() for s in structs)
    )
    with Popen('gofmt', stdin=PIPE, stdout=PIPE, stderr=PIPE) as proc:
        proc.stdin.write(package.encode('utf-8'))
        proc.stdin.close()
        err_out = proc.stderr.read()
        if err_out:
            print(err_out.decode('utf-8'), file=sys.stderr)
            sys.exit(1)
        output = proc.stdout.read()
        with open('./parser/ast_nodes.go', mode='wb') as fp:
            fp.write(output)


# ======================
# Put declarations here!
# ======================

if __name__ == '__main__':
    generate(
        # Statements
        stmts=[
            Struct('Module',   ['Filename string', 'Stmts []Stmt']),
            Struct('Let',      ['Name lexer.Token', 'Value Expr']),
            Struct('Block',    ['Stmts []Stmt']),
            Struct('For',      ['Keyword lexer.Token', 'Name lexer.Token', 'Iter Expr', 'Stmt Stmt']),
            Struct('While',    ['Cond Expr', 'Stmt Stmt']),
            Struct('If',       ['Cond Expr', 'Then Stmt', 'Else Stmt']),
            Struct('ExprStmt', ['Expr Expr']),
            Struct('Break',    ['Keyword lexer.Token']),
            Struct('Continue', ['Keyword lexer.Token']),
            Struct('Return',   ['Keyword lexer.Token', 'Expr Expr']),
        ],
        # Expressions
        exprs=[
            Struct('Binary',     ['Left Expr', 'Op lexer.Token', 'Right Expr']),
            Struct('And',        ['Left Expr', 'Op lexer.Token', 'Right Expr']),
            Struct('Or',         ['Left Expr', 'Op lexer.Token', 'Right Expr']),
            Struct('Assign',     ['Name lexer.Token', 'Right Expr'], extra_fields=['Loc int']),
            Struct('Unary',      ['Op lexer.Token', 'Right Expr']),
            Struct('Get',        ['Object Expr', 'Name lexer.Token']),
            Struct('Set',        ['Object Expr', 'Name lexer.Token', 'Right Expr']),
            Struct('Method',     ['Object Expr', 'Name lexer.Token', 'LParen lexer.Token', 'Args []Expr']),
            Struct('Call',       ['Callee Expr', 'LParen lexer.Token', 'Args []Expr']),
            Struct('GetIndex',   ['Left Expr', 'LBracket lexer.Token', 'Index Expr']),
            Struct('SetIndex',   ['Left Expr', 'LBracket lexer.Token', 'Index Expr', 'Right Expr']),
            Struct('Identifier', ['Id lexer.Token'], extra_fields=['Loc int']),
            Struct('Literal',    ['Lit lexer.Token']),
            Struct('Array',      ['Exprs []Expr']),
            Struct('Hash',       ['LBrace lexer.Token', 'Pairs []Pair']),
            Struct('Function',   ['Fn lexer.Token', 'Params []lexer.Token', 'Body *Block'], extra_fields=['Name string']),
            Struct('Super',      ['Tok lexer.Token', 'Name lexer.Token']),
        ],
    )
