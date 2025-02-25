def gather_user_profile(user_id: str):
    # Normal Python: fetch data from DB or API
    profile = {
        "user_id": user_id,
        "age": 30,
        "gender": "male",
        "allergies": ["nuts"]
    }
    return profile

def nutrition_flow(user_message: str, user_id: str):
    # 1) normal Python call
    profile = gather_user_profile(user_id)

    # 2) Store in context
    :context user_profile
    profile
    :--

    # 3) inline guardrail check
    :ask guardrail_check
    """
    Given user_message: {user_message}
    Return JSON { accept: bool, reason: str }
    Block if disallowed
    """
    :return_type "GuardrailOutput"
    :--

    # Now we can use normal Python to reference guardrail_check
    if not guardrail_check.accept:
        return f"BLOCKED: {guardrail_check.reason}"

    # 4) route decision
    :ask route
    """
    Decide route:
      - 'nutritionist' if user_message is about meals/diet
      - 'image_analysis' if user attached an image
      - 'support' otherwise
    Return JSON { route: string }
    """
    :return_type "RouteDecision"
    :--

    if route.route == "nutritionist":
        # 5) inline nutrition call
        :ask nutrition_block
        """
        Provide short dietary guidance using user_profile.
        Return JSON { accept: bool, reply: str }
        """
        :return_type "NutritionAdvice"
        :--
        return nutrition_block.reply

    elif route.route == "image_analysis":
        # Possibly do some Python logic to see if user provided an image
        has_image = False  # placeholder
        if not has_image:
            return "No image provided."

        :ask image_block
        """
        Analyze the image from user.
        Return JSON { accept: bool, data: object, reply: str }
        """
        :return_type "ImageAnalysisResult"
        :--

        if not image_block.accept:
            return image_block.reply
        
        # If recognized, do nutrition tracking
        :ask track_block
        """
        Combine recognized image data with user_profile
        Return JSON { reply: str }
        """
        :return_type "TrackingResult"
        :--

        return track_block.reply

    else:
        # 6) customer support
        :ask support_block
        """
        Provide basic account or site usage advice.
        Return JSON { accept: bool, reply: str }
        """
        :return_type "SupportResponse"
        :--
        return support_block.reply

# Call the function directly:
res = nutrition_flow("What's a good breakfast?", user_id="user123")
print("Result:", res)