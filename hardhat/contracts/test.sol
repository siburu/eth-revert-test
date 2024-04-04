//SPDX-License-Identifier: UNLICENSED

pragma solidity ^0.8.0;

error TestError(string msg, uint number);

contract C {
    C callee;

    enum RevertType { Require, Revert, CustomRevert }

    function setCallee(address _addr) external {
	callee = C(_addr);
    }

    function test(RevertType typ, uint depth) external {
	if (depth == 0) {
	    if (typ == RevertType.Require) {
		require(false, "test require");
	    } else if (typ == RevertType.Revert) {
		revert("test revert");
	    } else if (typ == RevertType.CustomRevert) {
		revert TestError("test custom error revert", 0xDEADBEEF);
	    } else {
		revert("unexpected RevertType");
	    }
	} else {
	    callee.test(typ, depth - 1);
	}
    }
}
