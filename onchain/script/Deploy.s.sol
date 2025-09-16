// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "forge-std/Script.sol";
import "../src/InferenceSettlement.sol";

contract Deploy is Script {
    function run() external {
        vm.startBroadcast();
        new InferenceSettlement();
        vm.stopBroadcast();
    }
}
