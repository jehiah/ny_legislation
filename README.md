A repository with information on New York State Legislation. This mirrors the [NY Senate Open Legislation Portal](https://legislation.nysenate.gov/).

## How can I use this data?

One way to do analysis on flat files is using tools like [jq](https://stedolan.github.io/jq/).

For example to see the top 10 most sponsored bills in each house which are not yet signed by the Govorner you could do the following

```
$ jq -r -c 'select(.Chamber == "SENATE" and .Resolution != true) | select(.Status != "SIGNED_BY_GOV") | select((.Sponsors | length) > 31) | {PrintNo,Status,Sponsors:(.Sponsors | length),Title}' bills/2023/*.json | jq -s -c 'sort_by(.Sponsors) | reverse | .[:10] | .[] '
{"PrintNo":"S8388","Status":"IN_SENATE_COMM","Sponsors":43,"Title":"Protects the health insurance benefits of retirees of public employers and contributions of retirees of public employers"}
{"PrintNo":"S6636","Status":"SENATE_FLOOR","Sponsors":43,"Title":"Provides for the types of damages that may be awarded to the persons for whose benefit an action for wrongful death is brought"}
{"PrintNo":"S3189","Status":"IN_SENATE_COMM","Sponsors":43,"Title":"Enacts the \"fair pay for home care act\""}
{"PrintNo":"S5085","Status":"SENATE_FLOOR","Sponsors":40,"Title":"Requires motor vehicle dealer franchisors to fully compensate franchised motor vehicle dealers for warranty service agreements"}
{"PrintNo":"S1292","Status":"IN_ASSEMBLY_COMM","Sponsors":40,"Title":"Establishes the clean fuel standard of 2024"}
{"PrintNo":"S7586","Status":"PASSED_ASSEMBLY","Sponsors":39,"Title":"Relates to permitting accidental death benefits to be awarded to the beneficiary of Anthony Varvaro"}
{"PrintNo":"S6880","Status":"IN_SENATE_COMM","Sponsors":36,"Title":"Relates to annual professional performance reviews of teachers and principals; repealer"}
{"PrintNo":"S568","Status":"IN_SENATE_COMM","Sponsors":36,"Title":"Relates to establishing the housing access voucher program"}
{"PrintNo":"S3397","Status":"PASSED_ASSEMBLY","Sponsors":36,"Title":"Establishes a maximum temperature in school buildings and indoor facilities"}
{"PrintNo":"S3170","Status":"IN_SENATE_COMM","Sponsors":36,"Title":"Establishes \"Kyra's Law\""}


$jq -r -c 'select(.Chamber == "ASSEMBLY" and .Resolution != true) | select(.Status != "SIGNED_BY_GOV") | select((.Sponsors | length) > 75) | {PrintNo,Status,Sponsors:(.Sponsors | length),Title}' bills/2023/*.json | jq -s -c 'sort_by(.Sponsors) | reverse | .[:10] | .[] '
{"PrintNo":"A8149","Status":"ASSEMBLY_FLOOR","Sponsors":109,"Title":"Establishes the New York child data protection act"}
{"PrintNo":"A8148","Status":"ASSEMBLY_FLOOR","Sponsors":108,"Title":"Establishes the Stop Addictive Feeds Exploitation (SAFE) for Kids act prohibiting the provision of addictive feeds to minors"}
{"PrintNo":"A7866","Status":"IN_ASSEMBLY_COMM","Sponsors":107,"Title":"Protects the health insurance benefits of retirees of public employers and contributions of retirees of public employers"}
{"PrintNo":"A1941","Status":"IN_ASSEMBLY_COMM","Sponsors":105,"Title":"Relates to providing universal school meals to students"}
{"PrintNo":"A964","Status":"IN_ASSEMBLY_COMM","Sponsors":101,"Title":"Establishes the clean fuel standard of 2024"}
{"PrintNo":"A2880","Status":"IN_ASSEMBLY_COMM","Sponsors":101,"Title":"Provides for paid family leave after a stillbirth"}
{"PrintNo":"A4066","Status":"PASSED_SENATE","Sponsors":99,"Title":"Requires motor vehicle dealer franchisors to fully compensate franchised motor vehicle dealers for warranty service agreements"}
{"PrintNo":"A3346","Status":"IN_ASSEMBLY_COMM","Sponsors":97,"Title":"Establishes \"Kyra's Law\""}
{"PrintNo":"A5990","Status":"ASSEMBLY_FLOOR","Sponsors":90,"Title":"Provides for the restriction of substances in menstrual products"}
{"PrintNo":"A5949","Status":"ASSEMBLY_FLOOR","Sponsors":87,"Title":"Prohibits the application of pesticides to certain local freshwater wetlands"}
```

