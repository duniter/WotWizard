/*
util: Set of tools.

Copyright (C) 2001-2020 Gérard Meunier

This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this program; if not, write to the Free Software Foundation, Inc., 59 Temple Place - Suite 330, Boston, MA 02111-1307, USA.
*/

package static

func fillArray () (compiler [1228]int32) {
	compiler = [1228]int32{
		   15,    35,     1,     2,     1,    32,    10,    21,     2,
		    1,     0,     0,     0,     0,     0,     0,     0,     1,
		    1,     1,     1,     1,     1,     1,     0,     1,     0,
		    1,     0,     1,     0,     1,     0,     1,     0,     1,
		    1,     1,     1,    13,   102,   105,   110,    32,   100,
		  101,    32,   116,   101,   120,   116,   101,     0,     0,
		    0,     0,     8,    34,   102,    97,   108,   115,   101,
		   34,     0,     7,    34,   110,   117,   108,   108,    34,
		    0,     7,    34,   116,   114,   117,   101,    34,     0,
		    4,    34,    44,    34,     0,     4,    34,    58,    34,
		    0,     4,    34,    91,    34,     0,     4,    34,    93,
		   34,     0,     4,    34,   123,    34,     0,     4,    34,
		  125,    34,     0,     7,   115,   116,   114,   105,   110,
		  103,     0,     7,   110,   117,   109,    98,   101,   114,
		    0,    15,    19,     0,    14,     4,     0,    14,     3,
		    0,    15,     2,     0,    15,     5,     0,    12,     0,
		    0,    11,     0,     0,    10,     0,     0,     9,     0,
		    0,     8,     0,     0,     7,     0,     0,    15,     1,
		    0,    15,     1,     0,    15,     1,     0,     3,     4,
		    0,     0,     0,     0,    14,     3,     0,    14,     3,
		    0,    13,     0,     0,    15,     9,     0,    15,     1,
		    0,    15,     1,     0,    15,     1,     0,    14,     1,
		    0,    15,     3,     0,    15,     1,     0,    15,     1,
		    0,    15,     1,     0,    15,     3,     0,     6,     0,
		    0,     5,     0,     0,    15,     1,     0,    15,     3,
		    0,     4,     0,     0,    15,     3,     0,    15,     0,
		    0,    14,     2,     2,    14,     9,    10,    14,    13,
		   13,    15,    26,    26,    14,    32,    32,     4,    34,
		   34,    10,    44,    44,     3,    45,    45,     2,    48,
		   48,     1,    49,    57,     9,    58,    58,     8,    91,
		   91,     7,    93,    93,    13,   102,   102,    12,   110,
		  110,    11,   116,   116,     6,   123,   123,     5,   125,
		  125,    17,    46,    46,     1,    48,    57,    16,    69,
		   69,    16,   101,   101,    17,    46,    46,    16,    69,
		   69,    16,   101,   101,     2,    48,    48,     1,    49,
		   57,     4,    32,    33,    18,    34,    34,     4,    35,
		   91,    19,    92,    92,     4,    93, 65535,    20,   114,
		  114,    21,   117,   117,    22,    97,    97,    14,     2,
		    2,    14,     9,    10,    14,    13,    13,    14,    32,
		   32,    23,    43,    43,    23,    45,    45,    23,    48,
		   57,    17,    48,    57,    16,    69,    69,    16,   101,
		  101,     4,    34,    34,     4,    47,    47,     4,    92,
		   92,     4,    98,    98,     4,   102,   102,     4,   110,
		  110,     4,   114,   114,     4,   116,   116,    24,   117,
		  117,    25,   117,   117,    26,   108,   108,    27,   108,
		  108,    23,    48,    57,    28,    48,    57,    28,    65,
		   70,    28,    97,   102,    29,   101,   101,    30,   108,
		  108,    31,   115,   115,    32,    48,    57,    32,    65,
		   70,    32,    97,   102,    33,   101,   101,    34,    48,
		   57,    34,    65,    70,    34,    97,   102,     4,    48,
		   57,     4,    65,    70,     4,    97,   102,     3,     2,
		    0,     0,     0,     0,     1,     0,     0,     1,    26,
		   26,     3,     9,     3,     2,     2,     2,     4,     4,
		    4,     4,     4,     2,     2,     2,     3,     2,     2,
		    3,     2,     2,     4,     4,     8,     2,     8,     2,
		    2,     3,     3,     3,     2,     2,     0,     9,     9,
		    1,     0,    11,    11,     2,     3,     0,    14,     0,
		    0,     4,     4,     6,     0,     5,     5,     7,     0,
		    6,     6,     8,     0,     9,     9,     1,     1,    10,
		   10,     5,     0,    11,    11,     2,     0,    13,    13,
		    9,     0,    14,    14,    10,     3,     0,    12,     0,
		    1,    12,    12,     9,     0,    13,    13,    15,     3,
		    0,    14,     0,     0,    10,    10,    18,     3,     0,
		   14,     0,     2,     0,     0,     0,     3,     0,    14,
		    0,     0,    12,    12,    19,     3,     0,    14,     0,
		    1,     7,     7,    15,     1,    10,    10,    15,     1,
		   12,    12,    15,     3,     0,    14,     0,     1,     7,
		    7,    14,     1,    10,    10,    14,     1,    12,    12,
		   14,     3,     0,    14,     0,     1,     7,     7,    16,
		    1,    10,    10,    16,     1,    12,    12,    16,     3,
		    0,    14,     0,     1,     7,     7,    20,     1,    10,
		   10,    20,     1,    12,    12,    20,     3,     0,    14,
		    0,     1,     7,     7,    19,     1,    10,    10,    19,
		    1,    12,    12,    19,     3,     0,    14,     0,     0,
		   10,    10,    20,     3,     0,    14,     0,     1,    10,
		   10,     1,     3,     0,    14,     0,     0,    12,    12,
		   21,     3,     0,    14,     0,     0,     7,     7,    22,
		    1,    10,    10,     3,     3,     0,    14,     0,     0,
		    8,     8,    24,     3,     0,    14,     0,     1,    12,
		   12,    12,     3,     0,    14,     0,     0,     7,     7,
		   25,     1,    12,    12,     7,     3,     0,    14,     0,
		    1,     0,     0,    10,     3,     0,    14,     0,     1,
		    0,     0,    11,     3,     0,    14,     0,     1,     7,
		    7,    17,     1,    10,    10,    17,     1,    12,    12,
		   17,     3,     0,    14,     0,     1,     7,     7,    18,
		    1,    10,    10,    18,     1,    12,    12,    18,     3,
		    0,    14,     0,     0,     4,     4,     6,     0,     5,
		    5,     7,     0,     6,     6,     8,     0,     9,     9,
		    1,     0,    11,    11,     2,     0,    13,    13,     9,
		    0,    14,    14,    10,     3,     0,    12,     0,     1,
		   10,    10,     4,     3,     0,    14,     0,     0,     4,
		    4,     6,     0,     5,     5,     7,     0,     6,     6,
		    8,     0,     9,     9,     1,     0,    11,    11,     2,
		    0,    13,    13,     9,     0,    14,    14,    10,     3,
		    0,    12,     0,     0,    13,    13,    15,     3,     0,
		   14,     0,     1,    12,    12,     8,     3,     0,    14,
		    0,     0,     7,     7,    22,     1,    10,    10,     3,
		    3,     0,    14,     0,     1,     7,     7,    13,     1,
		   12,    12,    13,     3,     0,    14,     0,     0,     7,
		    7,    25,     1,    12,    12,     7,     3,     0,    14,
		    0,     1,    10,    10,     2,     3,     0,    14,     0,
		    1,    12,    12,     6,     3,     0,    14,     0,     0,
		    0,     1,     2,     1,     2,     1,     1,     1,     2,
		    1,     1,     2,     1,     1,     2,     2,     2,     1,
		    3,     0,     0,     3,    24,    11,     0,    27,    30,
		   14,    23,     0,     1,    12,     0,    29,    31,    17,
		   26,     0,     2,    16,     0,     0,     0,     4,     0,
		    0,     5,    24,    13,     0,     0,    25,    29,     2,
		   17,     0,    22,    27,    24,    28,     1,    14,     1,
		    0,     0,     2,     1,     1,     3,     2,     1,     0,
		    2,     1,     2,     3,     1,     0,     3,     1,     3,
		    4,     1,     0,     4,     1,     2,     5,     1,     0,
		    5,     1,     2,     6,     2,     2,     6,     2,     2,
		    7,     1,     3,     8,     2,     1,     9,     1,     1,
		    9,     1,     1,     9,     1,     2,     9,     1,     2,
		    9,     1,     1,     9,     1,     1,     9,     1,     1,
		    0,     1,     0,     1,     1,     1,     1,     0,     1,
		    4,     2,     2,     1,     1,     1,     1,     0,     1,
		    3,     0,     1,     0,     1,     4,     2,     2,     1,
		    1,     1,     1,     0,     1,     3,     0,     1,     0,
		    1,     4,     3,     2,     1,     2,     2,     1,     1,
		    1,     0,     1,     3,     0,     1,     0,     1,     4,
		    3,     2,     1,     2,     2,     1,     1,     1,     0,
		    1,     3,     0,     0,     0,     2,     1,     1,     1,
		    0,     1,     2,     1,     0,     1,     2,     1,     0,
		    0,     2,     1,     1,     1,     0,     1,     1,     1,
		    0,     1,     2,     1,     1,     0,     1,     0,     1,
		    1,     1,     1,     0,     2,     0,     1,     1,     0,
		    1,     0,     1,     1,     1,     3,     1,     0,     0,
		    1,     6,     1,     1,     1,     0,     0,     1,     5,
		    1,     1,     1,     0,     0,     1,     4,     1,     1,
		    1,     1,     0,     1,     2,     1,     2,     1,     1,
		    0,     1,     1,     1,     2,     1,     0,     0,     1,
		    3,     1,     1,     1,     0,     0,     1,     2,     1,
		    1,     1,    10,     4,    12,     4,     0,     1,     1,
		    1,    22,     1,    24,     1,     0,     7,     1,     7,
		   22,     7,    24,     7,
	}
	return
}