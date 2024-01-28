import * as z from "zod"

export const SignupValidationSchema = z.object({
	username: z.string().min(2, { message: "Too short"}),
	password: z.string().min(8, { message: "Password must be at least 8 characters"})
})

