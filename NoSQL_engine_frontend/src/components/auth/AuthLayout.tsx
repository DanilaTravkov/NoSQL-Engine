import {Outlet, Navigate} from 'react-router-dom';

const AuthLayout = () => {

	const isAuthenticated = false

	return (
		<>
		{isAuthenticated ? (
			<Navigate to="/" />
		) : (
			<>
				<section className="flex flex-1 justify-center items-center flex-col py-10">
					<Outlet />
				</section>

				<div className='hidden xl:flex xl:flex-col xl:items-center xl:justify-center h-screen w-1/2 bg-black text-2xl'>
					<p className='pb-3'>NoSQL key-value engine </p>
					<img src="src/assets/software-svgrepo-com.svg" alt="logo" />
				</div>

				{/* <img src="src/assets/software.jpeg" alt="logo" className='hidden xl:block h-screen w-1/2 bg-no-repeat object-cover' /> */}
			</>
		)}
		</>
	)
}

export default AuthLayout