//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// Aliases
//

package netsim

import (
	"github.com/rbmk-project/rbmk/pkg/x/netsim/link"
	"github.com/rbmk-project/rbmk/pkg/x/netsim/netstack"
)

// Stack is an alias for [netstack.Stack].
type Stack = netstack.Stack

// Link is an alias for [link.Link].
type Link = link.Link

// NewStack is an alias for [netstack.New].
var NewStack = netstack.New

// NewLink is an alias for [link.New].
var NewLink = link.New
