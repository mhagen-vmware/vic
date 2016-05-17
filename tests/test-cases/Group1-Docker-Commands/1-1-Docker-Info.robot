*** Settings ***
Documentation  Test 1-1 - Docker Info
Resource  ../../resources/Util.robot
Test Setup  Install VIC Appliance To ESXi Server

*** Test Cases ***
Basic Info
    Log To Console  \nRunning docker info command...
    ${output}=  Run  docker ${params} info
    Log  ${output}
    Should contain  ${output}  Name: VIC

Debug Info
    ${status}=  Get State Of Github Issue  780
    Run Keyword If  '${status}' == 'closed'  Fail  Test test1.robot needs to be updated now that Issue #780 has been resolved
    #Log To Console  \nRunning docker -D info command...
    #${output}=  Run  docker ${params} -D info
    #Log  ${output}
    #Should contain  ${output}  Debug mode