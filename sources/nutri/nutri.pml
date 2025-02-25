:context user_demographics
  user_id = "123456789"
  name = "John"
  age = 30
  gender = "male"
  height_in_inches = 64
  weight_in_pounds = 140
  goal_weight = 130
  dietary_preferences = ["vegetarian", "low-sugar"]
  allergies = ["nuts"]
:--

:context meal_plan
  # A subset of the user’s recommended 7-day meal plan, as JSON or Python list
  # (We can store the entire structure or a pointer to it)
  recommended_plan = [
    {
      "date": "2025-02-25",
      "meals": [
        {"name": "Oatmeal with Berries", "meal_type": "breakfast", "calories": 350},
        ...
      ]
    },
    ...
  ]
:--

:plan "Nutritionist Copilot Flow"
summary:
  1. **Guardrail**: Evaluate user input for policy compliance.
  2. **Router**: If input is valid, route to:
     - “nutritionist” for meal queries
     - “image_analysis” if there's an attached image
     - “customer_support” for account help
  3. **Expert**: Based on route:
     - **NutritionExpert** provides advice
     - **ImageAnalysis** or **NutritionAnalysis** agent reviews images
     - **CustomerSupport** agent addresses user issues
:--

# A snippet to show how we might unify the guardrail + router steps
# Instead of calling them as big Python classes, we treat each as a directive.

:ask guardrail_check
"Inbound user message: {user_message}\n
Please decide whether to accept or block this message.
Output JSON with 'accept' (bool) and 'reply' (string if blocked)."
:return_type "GuardrailDecision"
:verify
  method = "json_schema"
  schema = {
    "accept": "boolean",
    "reply": "string"
  }
:--

#if the user_message is blocked:
:do
  if guardrail_check.accept == false:
    # We can store or log it, then return the 'reply' to the user
    # Possibly a new directive or a simple Python snippet:
    raise BlockedInputError(guardrail_check.reply)
:--

:ask route_decision
"User message: {user_message}\n
If guardrail_check.accept == true, decide the route:
  - 'nutritionist' if about meals, diet, or default
  - 'image_analysis' if image is attached
  - 'customer_support' if about account issues
Output JSON: { 'route': 'nutritionist' | 'image_analysis' | 'customer_support' }"
:return_type "RouteDecision"
:verify
  method = "json_schema"
  schema = {
    "route": ["nutritionist", "image_analysis", "customer_support"]
  }
:--

:do
  # Example pseudo-Python snippet for branching:
  if route_decision.route == "image_analysis":
      # Next directive calls image analysis
      set next_directive = ":ask image_analysis"
  elif route_decision.route == "customer_support":
      set next_directive = ":ask customer_support"
  else:
      set next_directive = ":ask nutrition_expert"
:--

:ask image_analysis
"Attached image: {base64_image_data}
User request: {user_message}
Please detect if there's a recognizable food item.
Output JSON with:
 accept: bool
 reply: string or null
 data: {some data about recognized food}
 route: 'nutritionist'"
:return_type "ImageAnalysisResult"
:verify
  method = "json_schema"
  schema = {
    "accept": "boolean",
    "reply": "string",
    "data": "object",
    "route": "string"
  }
:--

:ask nutrition_analysis
"User wants to track meal from the image. Provide detailed breakdown of the recognized meal:
 - ingredients with macros
 - total calories
Then forward to 'nutritionist' for final advice."
:return_type "NutritionAnalysisResult"
:--

:ask nutrition_expert
"""
Use the following context:
- user_demographics
- meal_plan
- Possibly image data or partial analysis from prior steps

Question: {user_message} 
If an image was recognized, incorporate that info (ingredients, etc.) 
Return a short text response, < 500 chars, about how to align with user goals.
"""
:return_type "NutritionResponse"
:verify
  method = "json_schema"
  schema = {
    "accept": "boolean",
    "reply": "string"
  }
:--

:ask customer_support
"User input: {user_message}
Focus on account settings or general site assistance. 
Keep it under 500 chars. 
If user tries to schedule medical appointments or discuss prescriptions, politely disclaim that we only support basic account help."
:return_type "CustomerSupportResponse"
:verify
  method = "json_schema"
  schema = {
    "accept": "boolean",
    "reply": "string"
  }
:--

:reflect
"Review entire flow. Are we caching results in .plcache? Do we store or log user messages? 
Ensure we unify results into a single conversation thread. 
Possible improvements:
1) Expand guardrail coverage
2) Add planned verification steps for nutritional data
3) Automatic re-ask if the route_decision is inconclusive
"
:--

