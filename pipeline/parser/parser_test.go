// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"reflect"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestParser(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected Stmts
		err      string
		fail     bool
	}{
		{
			name: "if-condition-list-paren2",
			in:   `if ((a==b) && (a==c)) || a==d { }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op: OR,
								LHS: &ParenExpr{
									Param: &ConditionalExpr{
										Op: AND,
										LHS: &ParenExpr{
											Param: &ConditionalExpr{
												Op:  EQEQ,
												LHS: &Identifier{Name: "a"},
												RHS: &Identifier{Name: "b"},
											},
										},
										RHS: &ParenExpr{
											Param: &ConditionalExpr{
												Op:  EQEQ,
												LHS: &Identifier{Name: "a"},
												RHS: &Identifier{Name: "c"},
											},
										},
									},
								},
								RHS: &ConditionalExpr{
									Op:  EQEQ,
									LHS: &Identifier{Name: "a"},
									RHS: &Identifier{Name: "d"},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "if-condition-list-paren",
			in:   `if (a==b) && (a==c) { }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op: AND,
								LHS: &ParenExpr{
									Param: &ConditionalExpr{
										Op:  EQEQ,
										LHS: &Identifier{Name: "a"},
										RHS: &Identifier{Name: "b"},
									},
								},
								RHS: &ParenExpr{
									Param: &ConditionalExpr{
										Op:  EQEQ,
										LHS: &Identifier{Name: "a"},
										RHS: &Identifier{Name: "c"},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "if-condition-list",
			in:   `if a==b && a==c { }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op: AND,
								LHS: &ConditionalExpr{
									Op:  EQEQ,
									LHS: &Identifier{Name: "a"},
									RHS: &Identifier{Name: "b"},
								},
								RHS: &ConditionalExpr{
									Op:  EQEQ,
									LHS: &Identifier{Name: "a"},
									RHS: &Identifier{Name: "c"},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "if-error-non-condition",
			in:   `if { }`,
			fail: true,
		},

		{
			name: "if-elif-error-non-condition",
			in: `
			if key=="11" {

			} elif {

			}`,
			fail: true,
		},

		{
			name: "if-elif-elif-error-non-condition",
			in: `
			if key=="11" {

			} elif key=="22" {

			} elif {

			}`,
			fail: true,
		},

		{
			name: "if-elif-else-expr",
			in: `
			if key=="11" {
				g1(arg)
			} elif key=="22" {
				g2(arg)
			} else {
				g3(arg)
			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g1",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "22"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g2",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
					Else: Stmts{
						&FuncStmt{
							Name:  "g3",
							Param: []Node{&Identifier{Name: "arg"}},
						},
					},
				},
			},
		},

		{
			name: "if-elif-expr",
			in: `
			if key=="11" {
				g1(arg)
			} elif key=="22" {
				g2(arg)
			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g1",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "22"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g2",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "if-elif-else-non-stmts",
			in: `
			if key=="11"{

			} elif key=="22" {

			} else {

			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
						},
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "22"},
							},
						},
					},
					Else: Stmts{},
				},
			},
		},

		{
			name: "if-elif-non-stmts",
			in: `
			if key=="11" {

			} elif key=="22" {

			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
						},
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "22"},
							},
						},
					},
				},
			},
		},

		{
			name: "if-expr-non-stmts",
			in:   `if key=="11" { }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
						},
					},
				},
			},
		},
		{
			name: "if-else-expr-non-stmts",
			in: `
			if key=="11" {

			} else {

			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
						},
					},
				},
			},
		},

		{
			name: "if-expr-non-stmts",
			in:   `if key=="11" { }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
						},
					},
				},
			},
		},

		{
			name: "if-else-expr",
			in:   `if key=="11" { g1(arg) g2(arg) } else { h(arg) }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g1",
									Param: []Node{&Identifier{Name: "arg"}},
								},
								&FuncStmt{
									Name:  "g2",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
					Else: Stmts{
						&FuncStmt{
							Name:  "h",
							Param: []Node{&Identifier{Name: "arg"}},
						},
					},
				},
			},
		},
		// {
		// 	name: "if-else-expr",
		// 	in:   `if match(_,"./*")=="11" { g1(arg) g2(arg) } else { h(arg) }`,
		// 	expected: Stmts{
		// 		&IfelseStmt{
		// 			IfList: IfList{
		// 				&IfExpr{
		// 					Condition: &ConditionalExpr{
		// 						Op:  EQEQ,
		// 						LHS: &Identifier{Name: "match(_, './*')"},
		// 						RHS: &StringLiteral{Val: "11"},
		// 					},
		// 					Stmts: Stmts{
		// 						&FuncStmt{
		// 							Name:  "g1",
		// 							Param: []Node{&Identifier{Name: "arg"}},
		// 						},
		// 						&FuncStmt{
		// 							Name:  "g2",
		// 							Param: []Node{&Identifier{Name: "arg"}},
		// 						},
		// 					},
		// 				},
		// 			},
		// 			Else: Stmts{
		// 				&FuncStmt{
		// 					Name:  "h",
		// 					Param: []Node{&Identifier{Name: "arg"}},
		// 				},
		// 			},
		// 		},
		// 	},
		// },

		{
			name: "if-else-expr-newline",
			in: `
			if key=="11" {
				g(arg)
			} else {
				h(arg)
			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
					Else: Stmts{
						&FuncStmt{
							Name:  "h",
							Param: []Node{&Identifier{Name: "arg"}},
						},
					},
				},
			},
		},

		{
			name: "if-nil",
			in:   `if abc == nil {}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "abc"},
								RHS: &NilLiteral{},
							},
						},
					},
				},
			},
		},

		{
			name: "if-expr",
			in:   `if key=="11" { g(arg) }`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "if-expr-newline",
			in: `
			if key=="11" {
				g(arg)
			}`,
			expected: Stmts{
				&IfelseStmt{
					IfList: IfList{
						&IfExpr{
							Condition: &ConditionalExpr{
								Op:  EQEQ,
								LHS: &Identifier{Name: "key"},
								RHS: &StringLiteral{Val: "11"},
							},
							Stmts: Stmts{
								&FuncStmt{
									Name:  "g",
									Param: []Node{&Identifier{Name: "arg"}},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "array-list in function arg value",
			in:   `f([1, 2.0, "3", true, false, nil, null, id_123])`,
			expected: Stmts{
				&FuncStmt{
					Name: "f",
					Param: []Node{
						FuncArgList{
							&NumberLiteral{IsInt: true, Int: 1},
							&NumberLiteral{Float: 2.0},
							&StringLiteral{Val: "3"},
							&BoolLiteral{Val: true},
							&BoolLiteral{Val: false},
							&NilLiteral{},
							&NilLiteral{},
							&Identifier{Name: "id_123"},
						},
					},
				},
			},
		},

		{
			name: "attr-expr in function arg value",
			in:   `f(arg=a.b.c)`,
			expected: Stmts{
				&FuncStmt{
					Name: "f",
					Param: []Node{
						&AssignmentStmt{
							LHS: &Identifier{Name: "arg"},
							RHS: &AttrExpr{
								Obj: &Identifier{Name: "a"},
								Attr: &AttrExpr{
									Obj: &AttrExpr{
										Obj:  &Identifier{Name: "b"},
										Attr: &Identifier{Name: "c"},
									},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "func_call_in_assignement_right",
			in:   `a = fn("a", true, a1=["b", 1.1])`,
			expected: Stmts{
				&AssignmentStmt{
					LHS: &Identifier{Name: "a"},
					RHS: &FuncStmt{
						Name: "fn",
						Param: []Node{
							&StringLiteral{Val: "a"},
							&BoolLiteral{Val: true},
							&AssignmentStmt{
								LHS: &Identifier{Name: "a1"},
								RHS: FuncArgList{
									&StringLiteral{Val: "b"},
									&NumberLiteral{Float: 1.1},
								},
							},
						},
					},
				},
			},
		},

		{
			name: "naming args",
			in:   `f(arg1=1, arg2=2)`,
			expected: Stmts{
				&FuncStmt{
					Name: "f",
					Param: []Node{
						&AssignmentStmt{
							LHS: &Identifier{Name: "arg1"},
							RHS: &NumberLiteral{IsInt: true, Int: 1},
						},

						&AssignmentStmt{
							LHS: &Identifier{Name: "arg2"},
							RHS: &NumberLiteral{IsInt: true, Int: 2},
						},
					},
				},
			},
		},

		{
			in:   `f(arg1>1, arg2=2)`,
			fail: true,
		},

		{
			name: "json attr",
			in:   `f(.[2].x[3])`,
			expected: Stmts{
				&FuncStmt{
					Name: `f`,
					Param: []Node{
						&AttrExpr{
							Obj: &IndexExpr{Index: []int64{2}},
							Attr: &IndexExpr{
								Obj:   &Identifier{Name: "x"},
								Index: []int64{3},
							},
						},
					},
				},
			},
		},

		{
			name: "multi-dim arr",
			in:   `f(x.y[2.5])`,
			fail: true,
		},

		{
			in: `f(x.y[1][2].z)`,
			expected: Stmts{
				&FuncStmt{
					Name: `f`,
					Param: []Node{
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj: &IndexExpr{
									Obj:   &Identifier{Name: "y"},
									Index: []int64{1, 2},
								},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},

		{
			name: "case:-multiple-functions",
			in: `f1()
		f2()
		f3()`,
			expected: Stmts{
				&FuncStmt{
					Name: `f1`,
				},
				&FuncStmt{
					Name: `f2`,
				},
				&FuncStmt{
					Name: `f3`,
				},
			},
		},

		{
			name: "embedded functions",
			in:   `f1(g(f2("abc"), 123), 1,2,3)`,
			// 函数参数不可以是函数
			fail: true,
		},

		{
			name: "case: attr syntax in function arg",
			in:   `avg(x.y.z, 1,2,3, p68, re("cd"), pqa)`,
			// 函数参数不可以 re()
			fail: true,
		},

		{
			name: "attr syntax with index syntax in function arg",
			in:   `json(_, x.y[1].z)`,
			expected: Stmts{
				&FuncStmt{
					Name: "json",
					Param: []Node{
						&Identifier{Name: "_"},
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj: &IndexExpr{
									Obj:   &Identifier{Name: "y"},
									Index: []int64{1},
								},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},

		{
			name: "simple attr syntax",
			in:   `json(_, x.y.z)`,
			expected: Stmts{
				&FuncStmt{
					Name: "json",
					Param: []Node{
						&Identifier{Name: "_"},
						&AttrExpr{
							Obj: &Identifier{Name: "x"},
							Attr: &AttrExpr{
								Obj:  &Identifier{Name: "y"},
								Attr: &Identifier{Name: "z"},
							},
						},
					},
				},
			},
		},

		{
			name: "simple attr syntax",
			in:   `match(_,"p([a-z]+)ch")`,
			expected: Stmts{
				&FuncStmt{
					Name: "match",
					Param: []Node{
						&Identifier{Name: "_"},
						&StringLiteral{Val: "p([a-z]+)ch"},
					},
				},
			},
		},

		{
			name: "many param",
			in:   `f(a, b, 1, 2, )`,
			expected: Stmts{
				&FuncStmt{
					Name: "f",
					Param: []Node{
						&Identifier{Name: "a"},
						&Identifier{Name: "b"},
						&NumberLiteral{IsInt: true, Int: 1},
						&NumberLiteral{IsInt: true, Int: 2},
					},
				},
			},
		},

		{
			name: `func-arg-with-multi-line-string`,
			in: `abc(x, """
this
is
multiline-string
""")`,
			expected: Stmts{
				&FuncStmt{
					Name: "abc",
					Param: []Node{
						&Identifier{Name: "x"},
						&StringLiteral{Val: `
this
is
multiline-string
`},
					},
				},
			},
		},

		{
			name: `func-func`,
			in:   `f1() f2()`,
			expected: Stmts{
				&FuncStmt{
					Name:  "f1",
					Param: nil,
				},
				&FuncStmt{
					Name:  "f2",
					Param: nil,
				},
			},
		},
	}

	// for idx := len(cases) - 1; idx >= 0; idx-- {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := ParsePipeline(tc.in)

			if !tc.fail {
				testutil.Ok(t, err)
			} else {
				t.Logf("expected error: %s", err)
				testutil.NotOk(t, err, "")
				return
			}

			var stmts Stmts

			switch v := node.(type) {
			case Stmts:
				stmts = v
			default:
				t.Errorf("should not been here, type: %s", reflect.TypeOf(v))
				return
			}

			if !tc.fail {
				var x, y string
				x = tc.expected.String()
				y = stmts.String()
				testutil.Ok(t, err)
				testutil.Equals(t, x, y)
				t.Logf("ok %s -> %s", tc.in, y)
			} else {
				t.Logf("%s -> expect fail: %v", tc.in, err)
				testutil.NotOk(t, err, "")
			}
		})
	}
}
