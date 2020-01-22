Elevator Project
================


Summary
-------
Create software for controlling `n` elevators working in parallel across `m` floors.


Main requirements
-----------------
Be reasonable: There may be semantic hoops that you can jump through to create something that is "technically correct". Do not hesitate to contact us if you feel that something is ambiguous or missing from these requirements.

### No orders are lost
 - Once the light on a hall call button (buttons for calling an elevator to that floor; top 6 buttons on the control panel) is turned on, an elevator should arrive at that floor
 > Master-slave topologi. Master styrer alle operasjoner, slaver lytter og utfører. Når slave får bestillingen utenfra (hall call): send melding til master om ny bestilling, master registerer den og sender melding til alle slaver om å skru på lys på den etasjen, samtidig som master bestemmer hvilken slave som skal betjene ordren. 
 - Similarly for a cab call (for telling the elevator what floor you want to exit at; front 4 buttons on the control panel), but only the elevator at that specific workspace should take the order
 > *Cab calls prioriteres over hall calls*. Cab calls skal kun behandles lokalt av aktuell slave, men slaven skal oppdatere master om denne bestillingen i tilfellet det er en aktuell bestilling mellom etasjen slaven er i nå og der den er på vei. Eks: slave i etasje 1, cab call til etasje 4, på veien får master en hall call om opp fra etasje 3, da skal slaven som fikk cab call til etasje 4 *også* behandle den bestillingen fra etasje 3. 
 - This means handling network packet loss, losing network connection entirely, software that crashes, and losing power - both to the elevator motor and the machine that controls the elevator
   - For cab orders, handling loss of power/software crash implies that the orders are executed once service is restored
   > Mulig løsning: ha et program (typ watchdog) som sjekker om heisprogrammet kjører, og hvis det ikke kjører spawner 

   > No network: `ifconfig eno1 | grep -e RUNNING` skal ikke returnere noe, men skal få match hvis nettverkstilkobling
   - The time used to detect these failures should be reasonable, ie. on the order of magnitude of seconds (not minutes)
   - Network packet loss is not an error, and can occur at any time
 - If the elevator is disconnected from the network, it should still serve all the currently active orders (ie. whatever lights are showing)
   - It should also keep taking new cab calls, so that people can exit the elevator even if it is disconnected from the network
   > Lokalmodus, se under. 
   - The elevator software should not require reinitialization (manual restart) after intermittent network or motor power loss
   > Alltid prøv å få kontakt med master: tråd som pinger master periodisk for å være sikker på man har kontakt med master. TODO: bestem intervall (forslag: 1 sec)
   > 
   > Slaver forteller backup om den skal bli master eller ikke. Hvis slave har kontakt med master og backup: ALL GOOD. Hvis slave har kontakt med master, men ikke backup: ALL GOOD. Hvis slave har kontakt med backup, men ikke master: Master er nede, skift til å lytte til backup. Hvis backup får kontakt med mastser: ALL GOOD. Hvis backup ikke får kontakt med master: se om backup får melding fra slavene om de får kontakt med master, hvis ikke må backup bli master. TODO: lag en mer oversiktlig tabell av dette
   > 
   > Power loss: timeout
   > 
   > MVP: Hardkode master og backup

### Multiple elevators should be more efficient than one
 - The orders should be distributed across the elevators in a reasonable way
   - Ex: If all three elevators are idle and two of them are at the bottom floor, then a new order at the top floor should be handled by the closest elevator (ie. neither of the two at the bottom).
 > Master sin delegeringsalgoritme skal (forhåpentligvis) ta seg av dette punktet. TODO: design denne algoritmen
 - You are free to choose and design your own "cost function" of some sort: Minimal movement, minimal waiting time, etc.
 > Minimal movement / shortest distance med hensyn på riktig retning 
 - The project is not about creating the "best" or "optimal" distribution of orders. It only has to be clear that the elevators are cooperating and communicating.
 
### An individual elevator should behave sensibly and efficiently
 - No stopping at every floor "just to be safe"
 - The hall "call upward" and "call downward" buttons should behave differently
   - Ex: If the elevator is moving from floor 1 up to floor 4 and there is a downward order at floor 3, then the elevator should not stop on its way upward, but should return back to floor 3 on its way down
   > Se minimal movement. Master bestemmer hvem som tar hvilken ordre 
 
### The lights and buttons should function as expected
 - The hall call buttons on all workspaces should let you summon an elevator
 > Master informerer hver slave om ny hall call når det skjer. Slave sender melding til master om ny hall call, master informerer alle slaver om denne hall call og bestemmer hvilken slave som skal behandle den. 
 - Under normal circumstances, the lights on the hall buttons should show the same thing on all workspaces 
   - Under circumstances with high packet loss, at least one light must work as expected
   > Se over
 - The cab button lights should not be shared between elevators
 > Master informerer kun om hall calls til andre slaver. Men slaver må informerer master om nye cab calls, men slaven styrer cab call lys selv. 
 - The cab and hall button lights should turn on as soon as is reasonable after the button has been pressed
   - Not ever turning on the button lights because "no guarantee is offered" is not a valid solution
   - You are allowed to expect the user to press the button again if it does not light up
 - The cab and hall button lights should turn off when the corresponding order has been serviced
 > Hall call: master får melding om at heis er ankommet til etasje og informerer alle heiser om at ordren er betjent og at de skal skru av lysene.
 > 
 > Cab call: behandles lokalt av slavene, informerer master om at den er behandlet (hvis det finnes en hall call må master vite at denne er behandlet og informere andre slaver). 
 - The "door open" lamp should be used as a substitute for an actual door, and as such should not be switched on while the elevator is moving
   - The duration for keeping the door open should be in the 1-5 second range
 > Være sikker på at dørene lukkes før heisen starter og at døren ikke åpnes mellom etasjene

 
Start with `1 <= n <= 3` elevators, and `m == 4` floors. Try to avoid hard-coding these values: You should be able to add a fourth elevator with no extra configuration, or change the number of floors with minimal configuration. You do, however, not need to test for `n > 3` and `m != 4`.


Unspecified behaviour
---------------------
Some things are left intentionally unspecified. Their implementation will not be tested, and are therefore up to you.

Which orders are cleared when stopping at a floor
 - You can clear only the orders in the direction of travel, or assume that everyone enters/exits the elevator when the door opens
 > Behandle kun i samme retning. Hvis hall call fra samme etasje som heisen er i nå har ikke retningen noe å si. 
 
How the elevator behaves when it cannot connect to the network (router) during initialization
 - You can either enter a "single-elevator" mode, or refuse to start
 > Lokalmodus, ta kun i mot cab calls og behandle tilfellet som om slaven mistet kontakt. Dvs. to modus: kommandoer fra master eller kun cab calls (se under)
 
How the hall (call up, call down) buttons work when the elevator is disconnected from the network
 - You can optionally refuse to take these new orders
 > Når ingen nettverkskontakt, behandle aktiv ordre fra master, og nekt å ta i mot ordre fra utsiden (hall calls). Fortsett å betjene ordre fra innsiden (cab calls). 
 
Stop button & obstruction switch are disabled
   - Their functionality (if/when implemented) is up to you.
   > Dette blir **bonus**. 

   
Permitted assumptions
---------------------

The following assumptions will always be true during testing:
 - At least one elevator is always working normally
 - No multiple simultaneous errors: Only one error happens at a time, but the system must still return to a fail-safe state after this error
   - (Recall that network packet loss is *not* an error in this context, and must be considered regardless of any other (single) error that can occur)
 - No network partitioning: There will never be a situation where there are multiple sets of two or more elevators with no connection between them
   
Additional resources
--------------------

Go to [the project resources repository](https://github.com/TTK4145/Project-resources) to find more resources for doing the project. This information is not required for the project, and is therefore maintained separately.