// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title Infrasound
 * @notice Infrasound keeps track of the total amount of ETH burned on L1 by both Transaction and Blobs.
 */
contract Infrasound {
    /**
     * @notice Total amount of ETH burned
     */
    uint256 public total;

    /**
     * @notice Mapping of timestamp to total burn.
     */
    mapping(uint64 => uint256) public reports;

    /**
     * @notice Enum for time intervals in seconds.
     */
    enum TimeInterval {
        FIVE_MINUTES,
        ONE_HOUR,
        ONE_DAY,
        SEVEN_DAYS,
        ONE_MONTH,
        ONE_YEAR,
        TWO_YEARS
    }

    /**
     * @notice Allows the system address to submit a report.
     *
     * @param _time L1 timestamp the report corresponds to.
     * @param _burn Amount of ETH burned in the block.
     */
    function report(uint64 _time, uint64 _burn) external {
        require(
            msg.sender == 0x83Eaca815B59ACCcA26Bf2d6c7560a7D415F9bd0,
            "Infrasound: reports can only be made from system address"
        );

        total += _burn;
        reports[_time] = total;
    }

    /**
     * @notice Tallies up the total burn since a given time interval.
     *
     * @param _interval The time interval to tally from.
     *
     * @return Total amount of ETH burned since the given time interval.
     */
    function tally(TimeInterval _interval) external view returns (uint256) {
        uint64 timeNow = uint64(block.timestamp);
        uint64 intervalSeconds;

        if (_interval == TimeInterval.FIVE_MINUTES) {
            intervalSeconds = 300;
        } else if (_interval == TimeInterval.ONE_HOUR) {
            intervalSeconds = 3600;
        } else if (_interval == TimeInterval.ONE_DAY) {
            intervalSeconds = 86400;
        } else if (_interval == TimeInterval.SEVEN_DAYS) {
            intervalSeconds = 604800;
        } else if (_interval == TimeInterval.ONE_MONTH) {
            intervalSeconds = 2592000;
        } else if (_interval == TimeInterval.ONE_YEAR) {
            intervalSeconds = 31536000;
        } else if (_interval == TimeInterval.TWO_YEARS) {
            intervalSeconds = 63072000;
        } else {
            revert("Infrasound: Invalid interval");
        }

        uint64 intervalTime = timeNow - intervalSeconds;

        // Checks if enough data for the specified interval exists
        if (reports[intervalTime] == 0) {
            revert("Infrasound: No data available for the specified interval");
        }

        return total - reports[intervalTime];
    }
}
